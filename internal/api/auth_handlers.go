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
	Role     string `json:"role"` // optional; defaults to "patient"
}

type tokenResponse struct {
	Token string `json:"token"`
}

func (d AuthDeps) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid json", nil)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Role = strings.TrimSpace(strings.ToLower(req.Role))
	if req.Email == "" || len(req.Password) < 8 {
		ErrorJSON(w, http.StatusBadRequest, "email/password invalid", "password must be at least 8 chars")
		return
	}
	if req.Role == "" {
		req.Role = "patient"
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "hash error", nil)
		return
	}

	u, err := d.Q.CreateUser(r.Context(), gen.CreateUserParams{
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
	})
	if err != nil {
		// TEMP DEBUG — surfaces constraint errors, etc.
		println("CreateUser error:", err.Error())
		ErrorJSON(w, http.StatusBadRequest, "create user failed", err.Error())
		return
	}

	tok, err := auth.SignJWT(d.Cfg.JWTSecret, d.Cfg.JWTIssuer, d.Cfg.JWTTTLMinutes, u.ID, u.Role)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "token error", nil)
		return
	}

	JSON(w, http.StatusCreated, tokenResponse{Token: tok})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (d AuthDeps) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid json", nil)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		ErrorJSON(w, http.StatusBadRequest, "email/password required", nil)
		return
	}

	u, err := d.Q.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		// Don’t leak whether the email exists
		ErrorJSON(w, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}
	if err := auth.CheckPassword(u.PasswordHash, req.Password); err != nil {
		ErrorJSON(w, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}

	tok, err := auth.SignJWT(d.Cfg.JWTSecret, d.Cfg.JWTIssuer, d.Cfg.JWTTTLMinutes, u.ID, u.Role)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "token error", nil)
		return
	}

	JSON(w, http.StatusOK, tokenResponse{Token: tok})
}
