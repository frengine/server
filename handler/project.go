package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/frengine/server/auth"
	"github.com/frengine/server/project"
	"github.com/gorilla/mux"
)

type ProjectListHandler struct {
	Deps
}

func (h ProjectListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ps, err := h.Deps.ProjectStore.Search()
	if err != nil && err != project.ErrNoFound {
		h.LogErr.Println(err)
		respond500(w, r)
		return
	}

	// TODO: Last modified

	respondSuccess(w, r, ps, time.Time{})
}

type ProjectGetHandler struct {
	Deps
}

func (h ProjectGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pid, _ := strconv.Atoi(mux.Vars(r)["id"])

	p, ok := mustFetchProject(w, r, h.Deps, pid)
	if !ok {
		return
	}

	respondSuccess(w, r, p, p.LastModified())
}

type ProjectCreateHandler struct {
	Deps
}

type createResponse struct {
	ProjectID int `json:"projectID"`
}

type createReq struct {
	Name   string `json:"name"`
	Author uint   `json:"author"`
}

func (h ProjectCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := createReq{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	u, err := getUserFromVars(r)
	if err != nil {
		h.LogErr.Println(err)
		respond500(w, r)
		return
	}

	if req.Author == 0 {
		req.Author = u.ID
	}

	if u.ID != req.Author {
		respondError(w, r, http.StatusForbidden, "forbidden")
		return
	}

	pid, err := h.Deps.ProjectStore.Create(req.Name, auth.User{ID: req.Author})
	if err != nil {
		h.LogErr.Println(err)
		respond500(w, r)
		return
	}

	respondSuccess(w, r, createResponse{pid}, time.Time{})
}

type ProjectUpdateHandler struct {
	Deps
}

type updateReq struct {
	Name   string `json:"name"`
	Author uint   `json:"author"`
}

func (h ProjectUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pid, _ := strconv.Atoi(mux.Vars(r)["id"])

	p, ok := mustFetchProject(w, r, h.Deps, pid)
	if !ok {
		return
	}

	if !mustBeLoggedInAs(w, r, h.Deps, p.Author.ID) {
		return
	}

	req := updateReq{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Author > 0 {
		p.Author.ID = req.Author
	}

	err := h.Deps.ProjectStore.Update(*p)
	if err != nil {
		if err == project.ErrInvalidAuthor {
			respondError(w, r, http.StatusBadRequest, "invalid author")
			return
		}
		h.LogErr.Println(err)
		respond500(w, r)
		return
	}

	respondSuccess(w, r, "success", time.Time{})
}

type ProjectDeleteHandler struct {
	Deps
}

func (h ProjectDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pid, _ := strconv.Atoi(mux.Vars(r)["id"])

	p, err := h.Deps.ProjectStore.FetchByID(pid)
	if err != nil {
		if err == project.ErrNoFound {
			respond404(w, r)
			return
		}
		h.LogErr.Println(err)
		respond500(w, r)
		return
	}

	u, err := getUserFromVars(r)
	if err != nil {
		h.LogErr.Println(err)
		respond500(w, r)
		return
	}

	if p.Author.ID != u.ID {
		respondError(w, r, http.StatusForbidden, "forbidden")
		return
	}

	err = h.Deps.ProjectStore.Delete(pid)
	if err != nil {
		fmt.Println(err)
	}

	respondSuccess(w, r, "success", time.Time{})
}
