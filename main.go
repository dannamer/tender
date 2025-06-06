package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"tender/internal/model"
	"tender/internal/storage"
)

type app struct {
	store *storage.Storage
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (a *app) listTenders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	serviceTypes := r.URL.Query()["service_type"]
	res := make([]model.Tender, 0)
	for _, t := range a.store.Data.Tenders {
		if len(serviceTypes) == 0 || contains(serviceTypes, t.ServiceType) {
			res = append(res, t)
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	writeJSON(w, http.StatusOK, res)
}

func (a *app) createTender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		ServiceType     string `json:"serviceType"`
		OrganizationID  string `json:"organizationId"`
		CreatorUsername string `json:"creatorUsername"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	t := model.Tender{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Description:     req.Description,
		ServiceType:     req.ServiceType,
		OrganizationID:  req.OrganizationID,
		CreatorUsername: req.CreatorUsername,
		Status:          "Created",
		Version:         1,
		CreatedAt:       time.Now().UTC(),
	}
	if err := a.store.AddTender(t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (a *app) userTenders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	res := make([]model.Tender, 0)
	for _, t := range a.store.Data.Tenders {
		if t.CreatorUsername == username {
			res = append(res, t)
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	writeJSON(w, http.StatusOK, res)
}

func (a *app) tenderStatus(w http.ResponseWriter, r *http.Request, id string) {
	tender, ok := a.store.GetTender(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]string{"status": tender.Status})
	case http.MethodPut:
		status := r.URL.Query().Get("status")
		if status == "" {
			http.Error(w, "missing status", http.StatusBadRequest)
			return
		}
		tender.History = append(tender.History, model.TenderVersion{
			Name:        tender.Name,
			Description: tender.Description,
			ServiceType: tender.ServiceType,
			Status:      tender.Status,
			Version:     tender.Version,
			CreatedAt:   tender.CreatedAt,
		})
		tender.Status = status
		tender.Version++
		if err := a.store.UpdateTender(tender); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, tender)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *app) tenderEdit(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tender, ok := a.store.GetTender(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		ServiceType *string `json:"serviceType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	tender.History = append(tender.History, model.TenderVersion{
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      tender.Status,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	})
	if req.Name != nil {
		tender.Name = *req.Name
	}
	if req.Description != nil {
		tender.Description = *req.Description
	}
	if req.ServiceType != nil {
		tender.ServiceType = *req.ServiceType
	}
	tender.Version++
	if err := a.store.UpdateTender(tender); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, tender)
}

func (a *app) tenderRollback(w http.ResponseWriter, r *http.Request, id string, ver int) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tender, ok := a.store.GetTender(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if ver < 1 || ver > len(tender.History) {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}
	snap := tender.History[ver-1]
	tender.History = append(tender.History, model.TenderVersion{
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      tender.Status,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	})
	tender.Name = snap.Name
	tender.Description = snap.Description
	tender.ServiceType = snap.ServiceType
	tender.Status = snap.Status
	tender.Version++
	if err := a.store.UpdateTender(tender); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, tender)
}

func (a *app) createBid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		TenderID    string `json:"tenderId"`
		AuthorType  string `json:"authorType"`
		AuthorID    string `json:"authorId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	b := model.Bid{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		TenderID:    req.TenderID,
		AuthorType:  req.AuthorType,
		AuthorID:    req.AuthorID,
		Status:      "Created",
		Version:     1,
		CreatedAt:   time.Now().UTC(),
	}
	if err := a.store.AddBid(b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (a *app) userBids(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	res := make([]model.Bid, 0)
	for _, b := range a.store.Data.Bids {
		if b.AuthorID == username {
			res = append(res, b)
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	writeJSON(w, http.StatusOK, res)
}

func (a *app) bidsForTender(w http.ResponseWriter, r *http.Request, tenderID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	res := make([]model.Bid, 0)
	for _, b := range a.store.Data.Bids {
		if b.TenderID == tenderID {
			res = append(res, b)
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	writeJSON(w, http.StatusOK, res)
}

func (a *app) bidStatus(w http.ResponseWriter, r *http.Request, id string) {
	bid, ok := a.store.GetBid(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]string{"status": bid.Status})
	case http.MethodPut:
		status := r.URL.Query().Get("status")
		if status == "" {
			http.Error(w, "missing status", http.StatusBadRequest)
			return
		}
		bid.History = append(bid.History, model.BidVersion{
			Name:        bid.Name,
			Description: bid.Description,
			Status:      bid.Status,
			Decision:    bid.Decision,
			Feedback:    bid.Feedback,
			Version:     bid.Version,
			CreatedAt:   bid.CreatedAt,
		})
		bid.Status = status
		bid.Version++
		if err := a.store.UpdateBid(bid); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, bid)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *app) bidEdit(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	bid, ok := a.store.GetBid(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	bid.History = append(bid.History, model.BidVersion{
		Name:        bid.Name,
		Description: bid.Description,
		Status:      bid.Status,
		Decision:    bid.Decision,
		Feedback:    bid.Feedback,
		Version:     bid.Version,
		CreatedAt:   bid.CreatedAt,
	})
	if req.Name != nil {
		bid.Name = *req.Name
	}
	if req.Description != nil {
		bid.Description = *req.Description
	}
	bid.Version++
	if err := a.store.UpdateBid(bid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, bid)
}

func (a *app) bidDecision(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	bid, ok := a.store.GetBid(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	decision := r.URL.Query().Get("decision")
	if decision == "" {
		http.Error(w, "missing decision", http.StatusBadRequest)
		return
	}
	bid.Decision = decision
	if err := a.store.UpdateBid(bid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, bid)
}

func (a *app) bidFeedback(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	bid, ok := a.store.GetBid(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	fb := r.URL.Query().Get("bidFeedback")
	if fb == "" {
		http.Error(w, "missing bidFeedback", http.StatusBadRequest)
		return
	}
	bid.Feedback = fb
	if err := a.store.UpdateBid(bid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, bid)
}

func (a *app) bidRollback(w http.ResponseWriter, r *http.Request, id string, ver int) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	bid, ok := a.store.GetBid(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if ver < 1 || ver > len(bid.History) {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}
	snap := bid.History[ver-1]
	bid.History = append(bid.History, model.BidVersion{
		Name:        bid.Name,
		Description: bid.Description,
		Status:      bid.Status,
		Decision:    bid.Decision,
		Feedback:    bid.Feedback,
		Version:     bid.Version,
		CreatedAt:   bid.CreatedAt,
	})
	bid.Name = snap.Name
	bid.Description = snap.Description
	bid.Status = snap.Status
	bid.Decision = snap.Decision
	bid.Feedback = snap.Feedback
	bid.Version++
	if err := a.store.UpdateBid(bid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, bid)
}

func (a *app) bidReviews(w http.ResponseWriter, r *http.Request, tenderID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	author := r.URL.Query().Get("authorUsername")
	res := a.store.ListReviews(tenderID, author)
	writeJSON(w, http.StatusOK, res)
}

func contains(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

func (a *app) tendersRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tenders/")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		switch parts[1] {
		case "status":
			a.tenderStatus(w, r, parts[0])
			return
		case "edit":
			a.tenderEdit(w, r, parts[0])
			return
		case "rollback":
			if len(parts) == 3 {
				v, _ := strconv.Atoi(parts[2])
				a.tenderRollback(w, r, parts[0], v)
				return
			}
		}
	}
	http.NotFound(w, r)
}

func (a *app) bidsRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/bids/")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		switch parts[1] {
		case "list":
			a.bidsForTender(w, r, parts[0])
			return
		case "status":
			a.bidStatus(w, r, parts[0])
			return
		case "edit":
			a.bidEdit(w, r, parts[0])
			return
		case "submit_decision":
			a.bidDecision(w, r, parts[0])
			return
		case "feedback":
			a.bidFeedback(w, r, parts[0])
			return
		case "rollback":
			if len(parts) == 3 {
				v, _ := strconv.Atoi(parts[2])
				a.bidRollback(w, r, parts[0], v)
				return
			}
		case "reviews":
			a.bidReviews(w, r, parts[0])
			return
		}
	}
	http.NotFound(w, r)
}

func main() {
	addr := os.Getenv("SERVER_ADDRESS")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}
	store, err := storage.New("data.json")
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	app := &app{store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/ping", pingHandler)
	mux.HandleFunc("/api/tenders", app.listTenders)
	mux.HandleFunc("/api/tenders/new", app.createTender)
	mux.HandleFunc("/api/tenders/my", app.userTenders)
	mux.HandleFunc("/api/tenders/", app.tendersRouter)

	mux.HandleFunc("/api/bids/new", app.createBid)
	mux.HandleFunc("/api/bids/my", app.userBids)
	mux.HandleFunc("/api/bids/", app.bidsRouter)

	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
