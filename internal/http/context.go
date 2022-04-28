package http

import (
	"net/http"
)

type ContextKey string

const serviceKey ContextKey = "Service"

func ServiceFromRequest(request *http.Request) string {
	ctx := request.Context()
	return ctx.Value(serviceKey).(string)
}
