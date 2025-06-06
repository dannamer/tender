package handler

import "net/http"
import "encoding/json"

func writeJSON(w http.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    if v != nil {
        _ = json.NewEncoder(w).Encode(v)
    }
}

