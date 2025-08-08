package handlers

import (
	"database/sql"
	"encoding/json"
  "github.com/gusplusbus/trustflow/auth_server/internal/auth"
  "github.com/gusplusbus/trustflow/auth_server/internal/database"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type userParams struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

func (cfg *APIConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if params.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	user, err := cfg.DBQueries.GetUser(r.Context(), params.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unable to get user", http.StatusInternalServerError)
		return
	}

	if err = auth.CheckPasswordHash(params.Password, user.HashedPassword); err != nil {
		http.Error(w, "Password Check Failed", http.StatusUnauthorized)
		return
	}

	tokenSecret := os.Getenv("JWT_SECRET")
	if tokenSecret == "" {
		http.Error(w, "JWT secret not configured", http.StatusInternalServerError)
		return
	}
	accessToken, err := auth.MakeJWT(user.ID, tokenSecret, time.Hour)
	if err != nil {
		http.Error(w, "Failed to create access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		http.Error(w, "Failed to create refresh token", http.StatusInternalServerError)
		return
	}

  _, err = cfg.DBQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
    Token:     refreshToken,
    CreatedAt: time.Now().UTC(),
    UpdatedAt: time.Now().UTC(),
    UserID:    user.ID,
    ExpiresAt: sql.NullTime{Time: time.Now().UTC().Add(60 * 24 * time.Hour), Valid: true},
    RevokedAt: sql.NullTime{Valid: false}, // NULL
  })
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to store refresh token: %v", err), http.StatusInternalServerError)
    return
  }

	resp := struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
    Token        string `json:"token"`
    RefreshToken string `json:"refresh_token"`
	}{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:  accessToken,
		RefreshToken: refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}


func (cfg *APIConfig) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    dbToken, err := cfg.DBQueries.GetUserIDFromRefreshToken(r.Context(), token)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if dbToken.ExpiresAt.Valid && time.Now().UTC().After(dbToken.ExpiresAt.Time) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if dbToken.RevokedAt.Valid {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    tokenSecret := os.Getenv("JWT_SECRET")
    if tokenSecret == "" {
        http.Error(w, "JWT secret not configured", http.StatusInternalServerError)
        return
    }

    accessToken, err := auth.MakeJWT(dbToken.UserID, tokenSecret, time.Hour)
    if err != nil {
        http.Error(w, "Failed to create access token", http.StatusInternalServerError)
        return
    }

    resp := map[string]string{"token": accessToken}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}



func (cfg *APIConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var params userParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if params.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := cfg.DBQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (cfg *APIConfig) HandleRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    err = cfg.DBQueries.RevokeRefreshToken(r.Context(), token)
    if err != nil {
        http.Error(w, "Failed to revoke token", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func (cfg *APIConfig) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
    // Get authenticated user ID
    userID, ok := GetUserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var params struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if params.Email == "" || params.Password == "" {
        http.Error(w, "Email and password are required", http.StatusBadRequest)
        return
    }

    hashedPassword, err := auth.HashPassword(params.Password)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }

    user, err := cfg.DBQueries.UpdateUser(r.Context(), database.UpdateUserParams{
        ID:             userID,
        Email:          params.Email,
        HashedPassword: hashedPassword,
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to update user: %v", err), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(user)
}

