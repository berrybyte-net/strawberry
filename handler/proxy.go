package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-redis/redis/v8"
)

type Proxy struct {
	Rcli *redis.Client
}

var _ http.Handler = (*Proxy)(nil)

func (h *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "strawberry")

	host := stripPort(r.Host)
	target, err := h.Rcli.Get(r.Context(), "strawberry:"+host).Result()
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusBadRequest)

		fmt.Fprintf(w, "host %q not configured in whitelist\n", host)
		return
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusBadRequest)

		fmt.Fprintf(w, "could not parse target host: %s\n", err)
		return
	}

	httputil.NewSingleHostReverseProxy(targetURL).ServeHTTP(w, r)
}
