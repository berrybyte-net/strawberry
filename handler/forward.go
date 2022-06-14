package handler

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/berrybyte-net/strawberry/repository"
	"github.com/go-chi/render"
)

type Forward struct {
	seedRepo     repository.Seed
	maxBodyBytes int
}

func NewForward(seedRepo repository.Seed, maxBodyBytes int) http.Handler {
	return &Forward{
		seedRepo:     seedRepo,
		maxBodyBytes: maxBodyBytes,
	}
}

func (h *Forward) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "strawberry")

	tgt, err := h.seedRepo.Seed(r.Context(), stripPort(r.Host))
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"message": "Could not get seed.",
		})
		return
	}

	tgtu, err := url.Parse(tgt)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"message": "Could not parse target URL.",
		})
		return
	}

	r.Body = io.NopCloser(io.LimitReader(r.Body, int64(h.maxBodyBytes)))

	httputil.NewSingleHostReverseProxy(tgtu).ServeHTTP(w, r)
}
