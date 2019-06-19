package handler

import (
	"encoding/json"
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

type listResponse struct {
	Projects []project.Project `json:"projects"`
}

func (h ProjectListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ps, err := h.Deps.ProjectStore.Search()
	if err != nil && err != project.ErrNoFound {
		h.LogErr.Println(err)
		respond500(w, r)
		return
	}

	// TODO: Last modified

	respondSuccess(w, r, listResponse{ps}, time.Time{})
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

	if u.ID != req.Author {
		respondError(w, r, http.StatusForbidden, "forbidden")
		return
	}

	pid, err := h.Deps.ProjectStore.Create(req.Name, auth.User{ID: u.ID})
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

	err = h.Deps.ProjectStore.Update(p)
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
