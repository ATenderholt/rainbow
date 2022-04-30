package http

import (
	"context"
	"github.com/ATenderholt/rainbow/internal/domain"
	"github.com/go-chi/chi/v5/middleware"
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

func motoMiddleware(motoService MotoService) func(http.Handler) http.Handler {
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
			r.Body = ioutil.NopCloser(body)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			authorization := r.Header.Get(Authorization)
			contentType := r.Header.Get(ContentType)
			target := r.Header.Get(AmzTarget)

			request := domain.MotoRequest{
				Service:       service,
				Method:        r.Method,
				Path:          r.URL.Path,
				Authorization: authorization,
				ContentType:   contentType,
				Target:        target,
				Payload:       payload.String(),
			}

			err := motoService.SaveRequest(context.Background(), request)
			if err != nil {
				logger.Errorf("Unable to persist Moto request: %v", err)
			}
		}

		return http.HandlerFunc(f)
	}
}
