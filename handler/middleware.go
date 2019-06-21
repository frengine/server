package handler

import (
	"net/http"
	"strings"
	"time"
)

type LoggerWare struct {
	Deps
}

func (mw LoggerWare) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tm := time.Now()

		lrw := &loggingResponseWriter{w, 0}

		next.ServeHTTP(lrw, r)

		diff := time.Since(tm)

		ip := strings.Split(r.RemoteAddr, ":")[0]

		ms := float64(diff / time.Millisecond)

		mw.LogInfo.Printf("%5.1fms %03d %15s %s", ms, lrw.statusCode, ip, r.RequestURI)
	})
}
