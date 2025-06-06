package handler

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"

    "tender/internal/service"
)

type BidHandler struct {
    svc *service.BidService
}

func NewBidHandler(s *service.BidService) *BidHandler {
    return &BidHandler{svc: s}
}

func (h *BidHandler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Post("/new", h.create)
    r.Get("/my", h.userBids)
    r.Route("/{id}", func(r chi.Router) {
        r.Get("/status", h.status)
        r.Put("/status", h.status)
        r.Patch("/edit", h.edit)
        r.Put("/submit_decision", h.decision)
        r.Put("/feedback", h.feedback)
        r.Put("/rollback/{version}", h.rollback)
    })
    r.Get("/{tenderId}/list", h.listTender)
    r.Get("/{tenderId}/reviews", h.reviews)
    return r
}

func (h *BidHandler) create(w http.ResponseWriter, r *http.Request) {
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
    b, err := h.svc.Create(req.Name, req.Description, req.TenderID, req.AuthorType, req.AuthorID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, b)
}

func (h *BidHandler) userBids(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    res := h.svc.UserBids(r.URL.Query().Get("username"))
    writeJSON(w, http.StatusOK, res)
}

func (h *BidHandler) listTender(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    tenderID := chi.URLParam(r, "tenderId")
    res := h.svc.ListForTender(tenderID)
    writeJSON(w, http.StatusOK, res)
}

func (h *BidHandler) status(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    switch r.Method {
    case http.MethodGet:
        for _, b := range h.svc.ListForTender("") {
            if b.ID == id {
                writeJSON(w, http.StatusOK, map[string]string{"status": b.Status})
                return
            }
        }
        http.NotFound(w, r)
    case http.MethodPut:
        status := r.URL.Query().Get("status")
        if status == "" {
            http.Error(w, "missing status", http.StatusBadRequest)
            return
        }
        bid, err := h.svc.UpdateStatus(id, status)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        writeJSON(w, http.StatusOK, bid)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

func (h *BidHandler) edit(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPatch {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    id := chi.URLParam(r, "id")
    var req struct {
        Name        *string `json:"name"`
        Description *string `json:"description"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    bid, err := h.svc.Edit(id, req.Name, req.Description)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, bid)
}

func (h *BidHandler) decision(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    id := chi.URLParam(r, "id")
    dec := r.URL.Query().Get("decision")
    if dec == "" {
        http.Error(w, "missing decision", http.StatusBadRequest)
        return
    }
    bid, err := h.svc.Decision(id, dec)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, bid)
}

func (h *BidHandler) feedback(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    id := chi.URLParam(r, "id")
    fb := r.URL.Query().Get("bidFeedback")
    if fb == "" {
        http.Error(w, "missing bidFeedback", http.StatusBadRequest)
        return
    }
    bid, err := h.svc.Feedback(id, fb)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, bid)
}

func (h *BidHandler) rollback(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    id := chi.URLParam(r, "id")
    ver, _ := strconv.Atoi(chi.URLParam(r, "version"))
    bid, err := h.svc.Rollback(id, ver)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, bid)
}

func (h *BidHandler) reviews(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    tenderID := chi.URLParam(r, "tenderId")
    author := r.URL.Query().Get("authorUsername")
    res := h.svc.Reviews(tenderID, author)
    writeJSON(w, http.StatusOK, res)
}

