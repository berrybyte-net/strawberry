package handler

import (
	"net/http"
	"net/url"

	"github.com/berrybyte-net/strawberry/repository"
	"github.com/go-chi/render"
)

type PutSeed struct {
	seedRepo repository.Seed
}

func NewPutSeed(seedRepo repository.Seed) http.Handler {
	return &PutSeed{
		seedRepo: seedRepo,
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

	tgtu, err := url.Parse(data.Target)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"message": "Could not parse target URL.",
		})
		return
	}

	if err := h.seedRepo.PutSeed(r.Context(), data.Name, tgtu.String()); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"message": "Could not put seed.",
		})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{
		"message": "Successfully put seed.",
	})
}
