package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type CreateProjectReq struct {
	Title                 string `json:"title"`
	Description           string `json:"description"`
	DurationEstimate      int    `json:"duration_estimate"`
	TeamSize              int    `json:"team_size"`
	ApplicationCloseTime  string `json:"application_close_time"` // ISO string
}

type ProjectResp struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	Description string  `json:"description"`
	DurationEstimate int `json:"duration_estimate"`
	TeamSize   int      `json:"team_size"`
	ApplicationCloseTime string `json:"application_close_time"`
}

func HandleCreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromCtx(r.Context())
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req CreateProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// TODO: validate server-side too, then insert into DB.
	// For now, return a dummy response so FE can proceed.
	resp := ProjectResp{
		ID:        "tmp-" + time.Now().Format("20060102150405"),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Title:     req.Title,
		Description: req.Description,
		DurationEstimate: req.DurationEstimate,
		TeamSize:  req.TeamSize,
		ApplicationCloseTime: req.ApplicationCloseTime,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
