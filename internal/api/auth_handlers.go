package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/justanamir/medappoint/internal/auth"
	"github.com/justanamir/medappoint/internal/config"
	"github.com/justanamir/medappoint/internal/db/gen"
)

type AuthDeps struct {
	Cfg config.Config
	Q   *gen.Queries
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // default patient
}
type tokenResponse struct {
	Token string `json:"token"`
}

func (d AuthDeps) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || len(req.Password) < 8 {
		http.Error(w, "email/password invalid", 400)
		return
	}
	if req.Role == "" {
		req.Role = "patient"
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "hash error", 500)
		return
	}

	u, err := d.Q.CreateUser(r.Context(), gen.CreateUserParams{
		Email: req.Email, PasswordHash: hash, Role: req.Role,
	})
	if err != nil {
		// TEMPORARY DEBUG LOG — safe; shows constraint/extension errors
		println("CreateUser error:", err.Error())
		http.Error(w, "create user failed", 400)
		return
	}

	tok, err := auth.SignJWT(d.Cfg.JWTSecret, d.Cfg.JWTIssuer, d.Cfg.JWTTTLMinutes, u.ID, u.Role)
	if err != nil {
		http.Error(w, "token error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokenResponse{Token: tok})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (d AuthDeps) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email/password required", 400)
		return
	}

	u, err := d.Q.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}
	if err := auth.CheckPassword(u.PasswordHash, req.Password); err != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}

	tok, err := auth.SignJWT(d.Cfg.JWTSecret, d.Cfg.JWTIssuer, d.Cfg.JWTTTLMinutes, u.ID, u.Role)
	if err != nil {
		http.Error(w, "token error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokenResponse{Token: tok})
}
