package handler

import (
	"net/http"
)

type LoggerWare struct {
	Deps Deps
}

func (mw LoggerWare) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		mw.Deps.LogInfo.Printf("%dms %s: %s", 1337, r.RemoteAddr, r.RequestURI)
	})
}
