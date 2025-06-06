package handler

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"

    "tender/internal/service"
)

type TenderHandler struct {
    svc *service.TenderService
}

func NewTenderHandler(s *service.TenderService) *TenderHandler {
    return &TenderHandler{svc: s}
}

func (h *TenderHandler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Get("/", h.list)
    r.Post("/new", h.create)
    r.Get("/my", h.userTenders)
    r.Route("/{id}", func(r chi.Router) {
        r.Get("/status", h.status)
        r.Put("/status", h.status)
        r.Patch("/edit", h.edit)
        r.Put("/rollback/{version}", h.rollback)
    })
    return r
}

func (h *TenderHandler) list(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    res := h.svc.List(r.URL.Query()["service_type"])
    writeJSON(w, http.StatusOK, res)
}

func (h *TenderHandler) create(w http.ResponseWriter, r *http.Request) {
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
    t, err := h.svc.Create(req.Name, req.Description, req.ServiceType, req.OrganizationID, req.CreatorUsername)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, t)
}

func (h *TenderHandler) userTenders(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    res := h.svc.UserTenders(r.URL.Query().Get("username"))
    writeJSON(w, http.StatusOK, res)
}

func (h *TenderHandler) status(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    switch r.Method {
    case http.MethodGet:
        tList := h.svc.List(nil)
        var status string
        for _, t := range tList {
            if t.ID == id {
                status = t.Status
                break
            }
        }
        writeJSON(w, http.StatusOK, map[string]string{"status": status})
    case http.MethodPut:
        status := r.URL.Query().Get("status")
        if status == "" {
            http.Error(w, "missing status", http.StatusBadRequest)
            return
        }
        tender, err := h.svc.UpdateStatus(id, status)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        writeJSON(w, http.StatusOK, tender)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

func (h *TenderHandler) edit(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPatch {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    id := chi.URLParam(r, "id")
    var req struct {
        Name        *string `json:"name"`
        Description *string `json:"description"`
        ServiceType *string `json:"serviceType"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    tender, err := h.svc.Edit(id, req.Name, req.Description, req.ServiceType)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, tender)
}

func (h *TenderHandler) rollback(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    id := chi.URLParam(r, "id")
    v, _ := strconv.Atoi(chi.URLParam(r, "version"))
    tender, err := h.svc.Rollback(id, v)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, tender)
}

