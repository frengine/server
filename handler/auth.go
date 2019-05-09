package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/frengine/server/auth"
)

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	Success bool `json:"success"`

	Token string `json:"token"`

	User auth.User `json:"user"`
}

type LoginHandler struct {
	Deps Deps
}

func (h LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var loginReq loginRequest

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&loginReq)
	if err != nil {
		h.Deps.LogErr.Println(err)
		respondError(w, r, http.StatusBadRequest, "cannot decode json as body")
		return
	}

	loginResp := loginResponse{}

	user, err := auth.CheckLogin(loginReq.Login, loginReq.Password)
	if user == (auth.User{}) {
		respondSuccess(w, r, loginResp, time.Time{})
		return
	}

	loginResp.Success = true
	loginResp.User = user
	loginResp.Token = generateToken(user)

	respondSuccess(w, r, loginResp, time.Time{})
}

type registerRequest struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	Password2 string `json:"password2"`
}

type registerResponse struct {
	Success bool `json:"success"`
}

type RegisterHandler struct {
	Deps Deps
}

func (h RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	respondSuccess(w, r, nil, time.Time{})
}

func generateToken(user auth.User) string {
	return base64.StdEncoding.EncodeToString([]byte(user.Name))
}
