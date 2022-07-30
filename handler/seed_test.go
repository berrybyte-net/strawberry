package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berrybyte-net/strawberry/store"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestPostSeed(t *testing.T) {
	w := httptest.NewRecorder()

	NewPostSeed(store.NewMemory()).ServeHTTP(
		w,
		httptest.NewRequest(
			http.MethodPost,
			"/seeds",
			bytes.NewBufferString(`{"name": "strawberry.amogus.systems", "target": "https://berrybyte.net"}`),
		),
	)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "{\"message\":\"Successfully put seed to store.\"}\n", string(body))
}

func TestDeleteSeed(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodDelete,
		"/seeds/strawberry.amogus.systems",
		nil,
	)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("name", "strawberry.amogus.systems")

	stor := store.NewMemory()
	stor.PutSeed(context.Background(), "strawberry.amogus.systems", "https://berrybyte.net")
	NewDeleteSeed(stor).ServeHTTP(
		w,
		r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx)),
	)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "{\"message\":\"Successfully deleted seed from store.\"}\n", string(body))
}
