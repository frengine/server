package handler

import (
	"encoding/json"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/frengine/server/auth"
)

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponseSuccess struct {
	Success bool `json:"success"`

	Token string `json:"token"`

	User auth.User `json:"user"`
}

type loginResponseError struct {
	Success bool `json:"success"`
}

type LoginHandler struct {
	Deps
}

func (h LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var loginReq loginRequest

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&loginReq)
	if err != nil {
		h.LogErr.Println(err)
		respondError(w, r, http.StatusBadRequest, "cannot decode json as body")
		return
	}

	user, err := h.UserStore.CheckLogin(loginReq.Login, loginReq.Password)
	if err == auth.ErrNoFound {
		loginResp := loginResponseError{}
		respondSuccess(w, r, loginResp, time.Time{})
		return
	}
	if err != nil {
		h.LogErr.Println(err)
		respondError(w, r, http.StatusInternalServerError, "cannot fetch from database")
		return
	}

	h.LogErr.Println(h.Cfg)
	token, err := generateToken(user, h.Cfg.JWTSecret)
	if err != nil {
		h.LogErr.Println(err)
		respondError(w, r, http.StatusInternalServerError, "")
		return
	}

	loginResp := loginResponseSuccess{
		Success: true,
		User:    user,
		Token:   token,
	}

	respondSuccess(w, r, loginResp, time.Time{})
}

type registerRequest struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	Password2 string `json:"password2"`
}

type registerResponse struct {
	Success bool `json:"success"`
}

type RegisterHandler struct {
	Deps
}

func (h RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&req)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, "cannot decode json as body")
		return
	}

	// User form entry errors.
	if len(req.Name) == 0 || len(req.Password) == 0 {
		respondError(w, r, http.StatusForbidden, "missing required fields")
		return
	}
	if req.Password != req.Password2 {
		respondError(w, r, http.StatusForbidden, "passwords don't match")
		return
	}
	/*if len(req.Password) < 8 {
		respondError(w, r, http.StatusForbidden, "longer password pls")
		return
	}*/

	err = h.UserStore.Register(req.Name, req.Password)
	if err == auth.ErrAlreadyExists {
		respondError(w, r, http.StatusForbidden, "name already exists")
		return
	}
	if err != nil {
		h.LogErr.Println(err)
		respondError(w, r, http.StatusBadRequest, "cannot fetch from database")
		return
	}

	resp := registerResponse{true}

	respondSuccess(w, r, resp, time.Time{})
}

func generateToken(user auth.User, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":   user.ID,
		"since": time.Now().Unix(),
	})

	return token.SignedString([]byte(secret))
}
