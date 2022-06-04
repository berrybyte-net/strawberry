package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/berrybyte-net/strawberry/config"
	"github.com/berrybyte-net/strawberry/handler"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/exp/slices"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info(
		"reading configuration from file",
		zap.String("config_path", "config.toml"),
	)
	cfg, err := config.ParseFile("config.toml")
	if err != nil {
		logger.Fatal(
			"could not read configuration from file",
			zap.Error(err),
		)
	}

	logger.Info(
		"creating redis client",
		zap.String("host", cfg.Redis.Host),
		zap.Int("port", cfg.Redis.Port),
	)
	rcli := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
	})
	if _, err := rcli.Ping(context.Background()).Result(); err != nil {
		logger.Fatal(
			"could not ping redis server",
			zap.Error(err),
		)
	}

	logger.Info(
		"creating certificate manager",
		zap.String("cert_directory", cfg.CertDirectory),
		zap.String("directory_url", cfg.ACME.DirectoryURL),
		zap.String("email", cfg.ACME.Email),
	)
	cm := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(cfg.CertDirectory),
		HostPolicy: func(ctx context.Context, host string) error {
			if _, err := rcli.Get(ctx, "strawberry:"+host).Result(); err != nil {
				return fmt.Errorf("host %q not configured in whitelist", host)
			}

			return nil
		},
		Client: &acme.Client{
			DirectoryURL: cfg.ACME.DirectoryURL,
			UserAgent:    "autocert",
		},
		Email: cfg.ACME.Email,
	}

	logger.Info("configuring http server")
	httpSrv := &http.Server{
		Addr:         ":http",
		Handler:      &handler.Redirect{},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			logger.Error(
				"could not listen and serve http",
				zap.Error(err),
			)
		}
	}()

	logger.Info("configuring https server")
	httpsSrv := &http.Server{
		Addr: ":https",
		Handler: &handler.Proxy{
			Rcli: rcli,
		},
		TLSConfig: &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return cm.GetCertificate(hello)
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
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	go func() {
		if err := httpsSrv.ListenAndServeTLS("", ""); err != nil {
			logger.Error(
				"could not listen and serve https",
				zap.Error(err),
			)
		}
	}()

	logger.Info("configuring rest api server")
	putRHostHandler := &handler.PutRHost{
		Rcli: rcli,
	}
	apiSrv := &http.Server{
		Addr: ":" + strconv.Itoa(cfg.API.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Server", "strawberry")

			auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(auth) != 2 ||
				auth[0] != "Bearer" ||
				!slices.Contains(cfg.API.AllowedHosts, stripPort(r.RemoteAddr)) ||
				auth[1] != cfg.API.Token {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.WriteHeader(http.StatusUnauthorized)

				fmt.Fprintln(w, "you are not authorized to access this endpoint")
				return
			}

			switch r.Method {
			case http.MethodPut:
				putRHostHandler.ServeHTTP(w, r)
			default:
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.WriteHeader(http.StatusNotFound)

				fmt.Fprintln(w, "404 page not found")
			}
		}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if cfg.API.UseSSL {
		apiSrv.TLSConfig = &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return cm.GetCertificate(hello)
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
		if err := apiSrv.ListenAndServeTLS("", ""); err != nil {
			logger.Fatal(
				"could not listen and serve api",
				zap.Error(err),
			)
		}
	}

	if err := apiSrv.ListenAndServe(); err != nil {
		logger.Fatal(
			"could not listen and serve api",
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
