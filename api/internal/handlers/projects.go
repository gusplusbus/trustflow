package handlers

import (
  "context"
  "encoding/json"
  "log"
  "net/http"
  projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
  "github.com/gusplusbus/trustflow/api/internal/clients"
)

type CreateProjectReq struct {
  Title                string `json:"title"`
  Description          string `json:"description"`
  DurationEstimate     int    `json:"duration_estimate"`
  TeamSize             int    `json:"team_size"`
  ApplicationCloseTime string `json:"application_close_time"`
}

func HandleCreateProject(w http.ResponseWriter, r *http.Request) {
  uid, ok := UserIDFromCtx(r.Context())
  if !ok || uid == "" {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
  }

  var req CreateProjectReq
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    log.Printf("[create] bad json: %v", err)
    http.Error(w, "Bad request", http.StatusBadRequest)
    return
  }

  log.Printf("[create] user=%s title=%q team=%d dur=%d close=%q",
    uid, req.Title, req.TeamSize, req.DurationEstimate, req.ApplicationCloseTime)

  cl := clients.ProjectClient()
  grpcResp, err := cl.CreateProject(context.Background(), &projectv1.CreateProjectRequest{
    UserId:               uid,
    Title:                req.Title,
    Description:          req.Description,
    DurationEstimate:     int32(req.DurationEstimate),
    TeamSize:             int32(req.TeamSize),
    ApplicationCloseTime: req.ApplicationCloseTime,
  })
  if err != nil {
    log.Printf("[create] gRPC error: %v", err)
    http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(grpcResp.Project); err != nil {
    log.Printf("[create] encode error: %v", err)
  }
}
