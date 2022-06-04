package handler

import (
	"net"
	"net/http"
)

type Redirect struct{}

var _ http.Handler = (*Redirect)(nil)

func (h *Redirect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "strawberry")
	w.Header().Set("Connection", "close")

	http.Redirect(w, r, "https://"+stripPort(r.Host)+r.URL.RequestURI(), http.StatusMovedPermanently)
}

// stripPort strips port from a network address of the form "host:port", "host%zone:port", "[host]:port" or
// "[host%zone]:port".
func stripPort(hostport string) string {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return hostport
	}

	return host
}
