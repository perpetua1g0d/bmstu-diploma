package mw

import (
	"net/http"
	"strconv"
	"time"
)

func BaseMetricsMiddleware(next http.HandlerFunc, serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{w, http.StatusOK, 0}

		next(rw, r)

		duration := float64(time.Since(start).Milliseconds())
		status := strconv.Itoa(rw.status)

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status, serviceName).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, serviceName).Observe(duration)

		httpRequestSize.WithLabelValues(r.Method, r.URL.Path, serviceName).Observe(float64(r.ContentLength))
		httpResponseSize.WithLabelValues(r.Method, r.URL.Path, serviceName).Observe(float64(rw.size))
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
