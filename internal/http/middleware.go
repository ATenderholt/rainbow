package http

import (
	"bytes"
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
			case "lambda", "s3", "sqs":
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

type SqsResponse struct {
	service    SqsService
	wrapped    http.ResponseWriter
	payload    string
	statusCode int
}

func NewSqsResponse(service SqsService, w http.ResponseWriter, payload string) SqsResponse {
	return SqsResponse{
		service: service,
		wrapped: w,
		payload: payload,
	}
}

func (s SqsResponse) Header() http.Header {
	return s.wrapped.Header()
}

// ReadFrom is necessary because response is being decorated; otherwise io.Copy
// will throw a short read error
func (s SqsResponse) ReadFrom(r io.Reader) (int64, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return int64(len(b)), err
	}

	n, err := s.Write(b)
	return int64(n), err
}

func (s SqsResponse) Write(bytes []byte) (int, error) {
	logger.Infof("Handling Elastic response, status code: %d", s.statusCode)
	// if error, just forward response
	if s.statusCode != 200 {
		return s.wrapped.Write(bytes)
	}

	action := s.service.ParseAction(s.payload)
	switch action {
	case "CreateQueue":
		err := s.service.SaveAttributes(s.payload)
		if err != nil {
			logger.Errorf("unable to save queue attributes: %v", err)
			s.WriteHeader(http.StatusInternalServerError)
			return s.wrapped.Write([]byte("unable to save queue attributes"))
		}
		return s.wrapped.Write(bytes)
	case "GetQueueAttributes":
		b, err := s.service.DecorateAttributes(s.payload, bytes)
		if err != nil {
			logger.Errorf("unable to decorate queue attributes: %v", err)
			s.WriteHeader(http.StatusInternalServerError)
			return s.wrapped.Write([]byte("unable to decorate queue attributes"))
		}
		return s.wrapped.Write(b)
	default:
		return s.wrapped.Write(bytes)
	}
}

func (s *SqsResponse) WriteHeader(statusCode int) {
	s.statusCode = statusCode
	s.wrapped.WriteHeader(statusCode)
}

func sqsMiddleware(sqsService SqsService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			service := ServiceFromRequest(r)
			if service != "sqs" {
				next.ServeHTTP(w, r)
				return
			}

			payload, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()

			if err != nil {
				logger.Errorf("unable to read body for sqs request: %v", err)
				http.Error(w, "unable to read body for sqs request", http.StatusBadRequest)
				return
			}

			r.Body = ioutil.NopCloser(bytes.NewReader(payload))

			ww := NewSqsResponse(sqsService, w, string(payload))

			next.ServeHTTP(&ww, r)
		}

		return http.HandlerFunc(f)
	}
}
