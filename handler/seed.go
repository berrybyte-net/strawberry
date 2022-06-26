package handler

import (
	"net/http"
	"net/url"

	"github.com/berrybyte-net/strawberry/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type PutSeed struct {
	stor store.Store
}

func NewPutSeed(stor store.Store) http.Handler {
	return &PutSeed{
		stor: stor,
	}
}

func (h *PutSeed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Name   string `json:"name"`
		Target string `json:"target"`
	}
	if err := render.DecodeJSON(r.Body, &data); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"message": "Malformed JSON body was sent.",
		})
		return
	}
	defer r.Body.Close()

	if _, err := h.stor.Seed(r.Context(), data.Name); err == nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{
			"message": "Given name is unavailable.",
		})
		return
	}

	tgtu, err := url.Parse(data.Target)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"message": "Could not parse target URL.",
		})
		return
	}

	if err := h.stor.PutSeed(r.Context(), data.Name, tgtu.String()); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"message": "Could not put seed to store.",
		})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{
		"message": "Successfully put seed to store.",
	})
}

type DeleteSeed struct {
	stor store.Store
}

func NewDeleteSeed(stor store.Store) http.Handler {
	return &DeleteSeed{
		stor: stor,
	}
}

func (h *DeleteSeed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParamFromCtx(r.Context(), "name")

	if _, err := h.stor.Seed(r.Context(), name); store.IsNoSeedFound(err) {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"message": "No matching seed could be found.",
		})
		return
	}

	if err := h.stor.DeleteSeed(r.Context(), name); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"message": "Could not delete seed from store.",
		})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{
		"message": "Successfully deleted seed from store.",
	})
}
