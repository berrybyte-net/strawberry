package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cm := &autocert.Manager{
		Cache:  autocert.DirCache("certs"),
		Prompt: autocert.AcceptTOS,
		Email:  "ssl@amogus.systems",
	}

	httpSrv := &http.Server{
		Addr: ":http",
		Handler: cm.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				logger.Error("could not split host and port from remote addr", zap.Error(err))
			}

			http.Redirect(w, r, "https://"+host+r.URL.RequestURI(), http.StatusMovedPermanently)
		})),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			logger.Error("could not listen and serve http", zap.Error(err))
		}
	}()

	httpsSrv := &http.Server{
		Addr: ":https",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, world!"))
		}),
		TLSConfig: &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				logger.Debug("getting certificate using client hello info", zap.String("server_name", hello.ServerName))
				return cm.GetCertificate(hello)
			},
		},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	httpsSrv.TLSConfig.NextProtos = append(httpsSrv.TLSConfig.NextProtos, acme.ALPNProto) // enable tls-alpn ACME challenges

	if err := httpsSrv.ListenAndServeTLS("", ""); err != nil {
		logger.Fatal("could not listen and serve https", zap.Error(err))
	}
}
