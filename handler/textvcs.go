package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type TestHandler struct {
	Deps
}

func (h TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	respondSuccess(w, r, "nice"+mux.Vars(r)["uid"], time.Time{})
}
