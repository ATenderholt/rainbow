package http

import (
	"context"
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
