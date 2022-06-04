package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
)

type PutRHost struct {
	Rcli *redis.Client
}

var _ http.Handler = (*PutRHost)(nil)

func (h *PutRHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Name   string `json:"name"`
		Target string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusBadRequest)

		fmt.Fprintln(w, "malformed json body was sent")
		return
	}
	defer r.Body.Close()

	if _, err := h.Rcli.Set(r.Context(), "strawberry:"+data.Name, data.Target, 0).Result(); err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusInternalServerError)

		fmt.Fprintf(w, "could not put rhost: %s\n", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "successfully put rhost")
}
