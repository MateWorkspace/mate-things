package presentationhttpproxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// NewMinioProxy reverse-proxies "/minio-proxy/<bucket>/<object>?<query>"
// requests to the real (private) MinIO endpoint, stripping the
// "/minio-proxy" prefix and forwarding the rest of the path and query
// string untouched.
//
// The Host header must be rewritten to the MinIO target host: presigned
// URL signatures include "host" as a signed header, computed against the
// endpoint the client was configured with, so presenting any other Host to
// MinIO makes every proxied request fail signature validation. MinIO
// remains the sole authority on whether a request's signature is valid -
// this proxy never re-signs or re-validates anything itself.
func NewMinioProxy(endpoint string, useSsl bool) http.Handler {
	scheme := "http"
	if useSsl {
		scheme = "https"
	}
	target := &url.URL{Scheme: scheme, Host: endpoint}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			setForwardedHeaders(req)

			// The presigned query string is MinIO's entire auth story here.
			// A client following a redirect from our JWT-protected admin API
			// may still be carrying its original "Authorization: Bearer ..."
			// header (same-host redirects don't get it stripped) - forwarding
			// that alongside the presigned signature makes MinIO reject the
			// request as ambiguous ("multiple authentication types").
			req.Header.Del("Authorization")

			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/minio-proxy")
		},
	}

	return proxy
}

// NewFrontendProxy reverse-proxies every non-API, non-minio-proxy request
// to the Next.js frontend, unmodified (no path rewriting) - equivalent to
// the "location /" block that used to live in nginx.conf.
func NewFrontendProxy(address string) http.Handler {
	target := &url.URL{Scheme: "http", Host: address}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			setForwardedHeaders(req)

			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
		},
	}

	return proxy
}

func setForwardedHeaders(req *http.Request) {
	if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		req.Header.Set("X-Real-IP", host)
		if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
			req.Header.Set("X-Forwarded-For", prior+", "+host)
		} else {
			req.Header.Set("X-Forwarded-For", host)
		}
	}

	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	req.Header.Set("X-Forwarded-Proto", scheme)
}
