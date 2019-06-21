package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/frengine/server/auth"
	"github.com/gorilla/mux"
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
		respondJSON(w, r, http.StatusUnauthorized, loginResp, time.Time{})
		return
	}
	if err != nil {
		h.LogErr.Println(err)
		respondError(w, r, http.StatusInternalServerError, "cannot fetch from database")
		return
	}

	token, err := generateToken(user, []byte(h.Cfg.JWTSecret))
	if err != nil {
		h.LogErr.Println(err)
		respondError(w, r, http.StatusInternalServerError, "")
		return
	}

	loginResp := loginResponseSuccess{
		Success: true,
		User:    user,
		Token:   base64.StdEncoding.EncodeToString([]byte(token)),
		//Token: token,
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
	if len(req.Password) < 8 {
		respondError(w, r, http.StatusForbidden, "longer password pls")
		return
	}

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

type Claims struct {
	UID uint `json:"uid"`
	jwt.StandardClaims
}

func generateToken(user auth.User, secret []byte) (string, error) {
	expirationTime := time.Now().Add(72 * time.Hour)
	claims := &Claims{
		UID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}

type AuthWare struct {
	Deps
}

func (mv AuthWare) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Authorization")

		parts := strings.Split(key, " ")
		key = parts[len(parts)-1]
		if key == "" {
			respondError(w, r, http.StatusUnauthorized, "Authentication header empty")
			return
		}

		keyStr, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "invalid token 5")
			mv.LogErr.Println(err)
			return
		}

		claims := &Claims{}

		tkn, err := jwt.ParseWithClaims(string(keyStr), claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(mv.Deps.Cfg.JWTSecret), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				respondError(w, r, http.StatusUnauthorized, "expired token")
				return
			}
			respondError(w, r, http.StatusBadRequest, "invalid token 1")
			mv.LogErr.Println(err)
			return
		}
		if !tkn.Valid {
			respondError(w, r, http.StatusUnauthorized, "invalid token 0")
			mv.LogErr.Println(err)
			return
		}

		mux.Vars(r)["uid"] = strconv.Itoa(int(claims.UID))

		next.ServeHTTP(w, r)
	})
}
