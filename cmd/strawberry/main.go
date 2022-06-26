package main

import (
	"context"
	"crypto/tls"
	"flag"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/berrybyte-net/strawberry/config"
	"github.com/berrybyte-net/strawberry/handler"
	"github.com/berrybyte-net/strawberry/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/exp/slices"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config.toml", "path to configuration file")
	flag.Parse()

	lgr, _ := zap.NewProduction()
	defer lgr.Sync()

	lgr.Info(
		"reading configuration from file",
		zap.String("config_path", cfgPath),
	)
	cfg, err := config.ParseFile(cfgPath)
	if err != nil {
		lgr.Fatal(
			"could not read configuration from file",
			zap.String("config_path", cfgPath),
			zap.Error(err),
		)
	}

	rcli := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
	})
	lgr.Info(
		"pinging redis server",
		zap.String("host", cfg.Redis.Host),
		zap.Int("port", cfg.Redis.Port),
	)
	if _, err := rcli.Ping(context.Background()).Result(); err != nil {
		lgr.Fatal(
			"could not ping redis server",
			zap.String("host", cfg.Redis.Host),
			zap.Int("port", cfg.Redis.Port),
			zap.Error(err),
		)
	}

	stor := store.NewRedis(rcli, cfg.Redis.Prefix)
	cmgr := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(cfg.CertDirectory),
		HostPolicy: func(ctx context.Context, host string) error {
			if _, err := stor.Seed(ctx, host); err != nil {
				return err
			}

			return nil
		},
		Client: &acme.Client{
			DirectoryURL: cfg.ACME.DirectoryURL,
			UserAgent:    "autocert",
		},
		Email: cfg.ACME.Email,
	}

	httpSrv := &http.Server{
		Addr:           ":80",
		Handler:        handler.NewRedirect(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}
	go func() {
		lgr.Info(
			"configuring http server",
			zap.String("host", "0.0.0.0"),
			zap.Int("port", 80),
		)
		if err := httpSrv.ListenAndServe(); err != nil {
			lgr.Fatal(
				"could not listen and serve http",
				zap.String("host", "0.0.0.0"),
				zap.Int("port", 80),
				zap.Error(err),
			)
		}
	}()

	http2Srv := &http.Server{
		Addr:    ":443",
		Handler: handler.NewForward(stor, cfg.MaxBodyBytes, cfg.StrictSNIHost),
		TLSConfig: &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return cmgr.GetCertificate(hello)
			},
			NextProtos: []string{
				// Enable HTTP/2
				"h2",
				"http/1.1",
				// Enable TLS-ALPN ACME challenges
				acme.ALPNProto,
			},
			// https://blog.cloudflare.com/exposing-go-on-the-internet/
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			},
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
			},
		},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}
	go func() {
		lgr.Info(
			"configuring http2 server",
			zap.String("host", "0.0.0.0"),
			zap.Int("port", 443),
		)
		if err := http2Srv.ListenAndServeTLS("", ""); err != nil {
			lgr.Fatal(
				"could not listen and serve http2",
				zap.String("host", "0.0.0.0"),
				zap.Int("port", 443),
				zap.Error(err),
			)
		}
	}()

	apiRtr := chi.NewRouter()
	apiRtr.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(auth) != 2 ||
				auth[0] != "Bearer" ||
				!slices.Contains(cfg.API.AllowedIPs, stripPort(r.RemoteAddr)) ||
				cfg.API.Token == "" ||
				auth[1] != cfg.API.Token {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{
					"message": "Unauthorized to access this endpoint.",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	})
	apiRtr.Delete("/seeds/{name}", handler.NewDeleteSeed(stor).ServeHTTP)
	apiRtr.Put("/seeds", handler.NewPutSeed(stor).ServeHTTP)

	apiSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.API.Port),
		Handler:      apiRtr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if cfg.API.UseSSL {
		apiSrv.TLSConfig = &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return cmgr.GetCertificate(hello)
			},
			NextProtos: []string{
				// Enable HTTP/2
				"h2",
				"http/1.1",
				// Enable TLS-ALPN ACME challenges
				acme.ALPNProto,
			},
			// https://blog.cloudflare.com/exposing-go-on-the-internet/
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			},
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
			},
		}
		lgr.Info(
			"configuring api server",
			zap.String("host", "0.0.0.0"),
			zap.Int("port", cfg.API.Port),
		)
		if err := apiSrv.ListenAndServeTLS("", ""); err != nil {
			lgr.Fatal(
				"could not listen and serve api",
				zap.String("host", "0.0.0.0"),
				zap.Int("port", cfg.API.Port),
				zap.Error(err),
			)
		}
	}
	if err := apiSrv.ListenAndServe(); err != nil {
		lgr.Fatal(
			"could not listen and serve api",
			zap.String("host", "0.0.0.0"),
			zap.Int("port", cfg.API.Port),
			zap.Error(err),
		)
	}
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
