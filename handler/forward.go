package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/berrybyte-net/strawberry/store"
	"github.com/go-chi/render"
)

type Forward struct {
	stor          store.Store
	maxBodyBytes  int
	strictSNIHost bool
}

func NewForward(stor store.Store, maxBodyBytes int, strictSNIHost bool) http.Handler {
	return &Forward{
		stor:          stor,
		maxBodyBytes:  maxBodyBytes,
		strictSNIHost: strictSNIHost,
	}
}

func (h *Forward) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "strawberry")

	if r.TLS != nil && h.strictSNIHost && r.TLS.ServerName != stripPort(r.Host) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"message": "Request host and TLS ServerName values differ.",
		})
		return
	}

	tgt, err := h.stor.Seed(r.Context(), stripPort(r.Host))
	if err != nil {
		if store.IsNoSeedFound(err) {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{
				"message": "No matching seed could be found.",
			})
			return
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"message": "Could not get seed from store.",
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

	r.Body = http.MaxBytesReader(w, r.Body, int64(h.maxBodyBytes))
	httputil.NewSingleHostReverseProxy(tgtu).ServeHTTP(w, r)
}
