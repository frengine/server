package handler

import (
	"net/http"
	"time"
)

type LoggerWare struct {
	Deps Deps
}

func (mw LoggerWare) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tm := time.Now()

		next.ServeHTTP(w, r)

		diff := time.Since(tm)

		ms := float64(diff / time.Millisecond)

		mw.Deps.LogInfo.Printf("%.2fms %s: %s", ms, r.RemoteAddr, r.RequestURI)
	})
}
