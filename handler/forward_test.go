package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/berrybyte-net/strawberry/store"
	"github.com/go-chi/render"
	"github.com/stretchr/testify/assert"
)

func TestForward(t *testing.T) {
	tgts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]string{
			"message": "Hello, world!",
		})
	}))
	defer tgts.Close()

	stor := store.NewMemory()
	// 100MB
	fs := httptest.NewServer(NewForward(stor, 100000000, false, false))
	defer fs.Close()

	stor.PutSeed(context.Background(), stripPort(strings.TrimPrefix(fs.URL, "http://")), tgts.URL)

	resp, err := http.Get(fs.URL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "{\"message\":\"Hello, world!\"}\n", string(body))
}
