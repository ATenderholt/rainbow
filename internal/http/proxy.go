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

	url := r.URL
	url.Scheme = "http"
	switch service {
	case "lambda":
		url.Host = p.cfg.FunctionsHost()
	case "s3":
		url.Host = p.cfg.StorageHost()
	case "sqs":
		url.Host = p.cfg.SqsHost()
	default:
		url.Host = p.cfg.MotoHost()
	}

	logger.Infof("Proxying request for service %s to %s", service, url.String())

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

	client := &http.Client{}
	proxyRes, err := client.Do(proxyReq)
	if err != nil {
		logger.Errorf("Unable to make proxy request for service %s: %v", service, err)
		http.Error(w, "Unable to make proxy request", http.StatusInternalServerError)
		return
	}

	for key, value := range proxyRes.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
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
