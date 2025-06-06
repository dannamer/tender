package handler

import "net/http"

func Ping(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("ok"))
}

