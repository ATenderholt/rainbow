package http

import (
	"github.com/ATenderholt/rainbow/settings"
	"io"
	"net/http"
)

type Proxy struct {
	cfg *settings.Config
}

func NewProxy(cfg *settings.Config) Proxy {
	return Proxy{
		cfg: cfg,
	}
}

func (p Proxy) handle(w http.ResponseWriter, r *http.Request) {
	service := ServiceFromRequest(r)
	logger.Infof("Handling %s for service %s", r.URL.Path, service)

	url := r.URL
	url.Host = "localhost:5001"
	url.Scheme = "http"

	proxyReq, err := http.NewRequest(r.Method, url.String(), r.Body)
	if err != nil {
		logger.Errorf("Unable to create proxy request for service %s: %v", service, err)
		http.Error(w, "Unable to create proxy request", http.StatusInternalServerError)
		return
	}

	proxyReq.Header.Set("Host", r.Host)
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

	for header, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	logger.Infof("Proxying request for service %s to %s", service, url.String())

	client := &http.Client{}
	proxyRes, err := client.Do(proxyReq)
	if err != nil {
		logger.Errorf("Unable to make proxy request for service %s: %v", service, err)
		http.Error(w, "Unable to make proxy request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(proxyRes.StatusCode)
	_, err = io.Copy(w, proxyRes.Body)
	if err != nil {
		logger.Errorf("Unable to copy proxy response for service %s: %v", service, err)
		http.Error(w, "Unable to copy proxy response", http.StatusInternalServerError)
		return
	}

	defer proxyRes.Body.Close()
}
