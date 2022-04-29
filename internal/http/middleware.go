package http

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const (
	Authorization = "Authorization"
	ContentType   = "Content-Type"
	AmzTarget     = "X-Amz-Target"
)

var authRegex *regexp.Regexp

func init() {
	temp, err := regexp.Compile(`Credential=(\w+)/\d{8}/([a-z0-9-]+)/(\w+)/aws4_request`)
	if err != nil {
		panic(err)
	}

	authRegex = temp
}

func extractService(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		groups := authRegex.FindStringSubmatch(auth)
		var service string
		if groups == nil {
			logger.Errorf("Unable to match Authorization header: %s", auth)
			service = ""
		} else {
			service = groups[3]
		}

		ctx := context.WithValue(r.Context(), serviceKey, service)
		r = r.Clone(ctx)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func motoMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			service := ServiceFromRequest(r)
			switch service {
			case "lambda", "s3":
				next.ServeHTTP(w, r)
				return
			}

			var payload strings.Builder
			body := io.TeeReader(r.Body, &payload)
			//authorization := r.Header.Get(Authorization)
			//contentType := r.Header.Get(ContentType)
			//target := r.Header.Get(AmzTarget)

			r.Body = ioutil.NopCloser(body)

			next.ServeHTTP(w, r)

			logger.Infof("Back from %s after sending %s", r.RequestURI, payload.String())
		}

		return http.HandlerFunc(f)
	}
}
