package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/frengine/server/auth"
	"github.com/frengine/server/config"
	"github.com/frengine/server/project"
	"github.com/gorilla/mux"
)

type Deps struct {
	UserStore    auth.Store
	ProjectStore project.Store
	LogInfo      *log.Logger
	LogErr       *log.Logger
	Cfg          config.Config
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func respondJSON(w http.ResponseWriter, r *http.Request, code int, v interface{}, lm time.Time) (error, bool) {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err, false
	}

	// Manage If-Modified-Since and add Last-Modified.
	if lm != (time.Time{}) {
		if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && lm.Unix() <= t.Unix() {
			w.WriteHeader(http.StatusNotModified)
			return nil, false
		}
		w.Header().Set("Last-Modified", lm.Format(http.TimeFormat))
	}

	w.WriteHeader(code)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(data)
	return err, true
}

func respondError(w http.ResponseWriter, r *http.Request, code int, message string) (error, bool) {
	return respondJSON(w, r, code, map[string]string{"error": message}, time.Time{})
}

func respondSuccess(w http.ResponseWriter, r *http.Request, v interface{}, lm time.Time) (error, bool) {
	return respondJSON(w, r, http.StatusOK, v, lm)
}

func respond500(w http.ResponseWriter, r *http.Request) error {
	err, _ := respondError(w, r, http.StatusInternalServerError, "internal server error")
	return err
}

func respond404(w http.ResponseWriter, r *http.Request) error {
	err, _ := respondError(w, r, http.StatusNotFound, "not found")
	return err
}

func getUserFromVars(r *http.Request) (auth.User, error) {
	vars := mux.Vars(r)
	uid, _ := strconv.Atoi(vars["uid"])
	if uid == 0 {
		return auth.User{}, fmt.Errorf("cannot get uid from mux vars, is 0")
	}

	return auth.User{ID: uint(uid)}, nil
}
