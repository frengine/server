package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Deps struct {
	DB      *sql.DB
	LogInfo *log.Logger
	LogErr  *log.Logger
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
