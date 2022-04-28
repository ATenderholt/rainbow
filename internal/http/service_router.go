package http

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"regexp"
)

var authRegex *regexp.Regexp

func init() {
	temp, err := regexp.Compile(`Credential=(\w+)/\d{8}/([a-z0-9-]+)/(\w+)/aws4_request`)
	if err != nil {
		panic(err)
	}

	authRegex = temp
}

type ServiceRouter struct {
	IAM chi.Router
	STS chi.Router
}

func NewServiceRouter() *ServiceRouter {
	mux := chi.NewMux()

	sts := mux.With(stsMiddleware)
	sts.HandleFunc("/*", defaultHandler)

	iam := mux.With(iamMiddleware)
	iam.HandleFunc("/*", defaultHandler)

	return &ServiceRouter{
		IAM: iam,
		STS: sts,
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	logger.Infof("Handling %s for service %s", r.URL.Path, ServiceFromRequest(r))
}

func iamMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		logger.Info("IAM middleware")
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func stsMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		logger.Info("STS middleware")
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func (s ServiceRouter) RouteByAuthorization(w http.ResponseWriter, r *http.Request) {
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

	switch service {
	case "iam":
		s.IAM.ServeHTTP(w, r)
	case "sts":
		s.STS.ServeHTTP(w, r)
	default:
		http.HandlerFunc(defaultHandler).ServeHTTP(w, r)
	}

}
