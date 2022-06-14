package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berrybyte-net/strawberry/repository"
	"github.com/stretchr/testify/assert"
)

func TestPutSeed(t *testing.T) {
	w := httptest.NewRecorder()

	NewPutSeed(repository.NewMemory()).ServeHTTP(
		w,
		httptest.NewRequest(
			http.MethodPut,
			"/",
			bytes.NewBufferString(`{"name": "strawberry.amogus.systems", "target": "https://berrybyte.net"}`),
		),
	)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "{\"message\":\"Successfully put seed.\"}\n", string(body))
}
