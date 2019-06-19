package handler

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/frengine/server/project"
	"github.com/gorilla/mux"
)

type RevisionGetHandler struct {
	Deps
}

func (h RevisionGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pid, _ := strconv.Atoi(mux.Vars(r)["id"])

	rev, err := h.Deps.ProjectStore.FetchLatestRevisionByProject(pid)
	if err != nil {
		h.LogErr.Println(err)
		respond500(w, r)
		return
	}

	lm := time.Time{}
	if rev.Created != nil {
		lm = *rev.Created
	}

	respondSuccess(w, r, rev, lm)
}

type RevisionSaveHandler struct {
	Deps
}

func (h RevisionSaveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pid, _ := strconv.Atoi(mux.Vars(r)["id"])

	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.LogErr.Println(err)
		respond500(w, r)
		return
	}

	err = h.Deps.ProjectStore.SaveRevision(pid, string(content))
	if err != nil {
		if err == project.ErrInvalidProject {
			respondError(w, r, http.StatusBadRequest, "invalid project")
			return
		}
		h.LogErr.Println(err)
		respond500(w, r)
		return
	}

	respondSuccess(w, r, "succes", time.Time{})
}
