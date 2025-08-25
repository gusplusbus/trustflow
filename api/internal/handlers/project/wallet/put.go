package wallet


import (
  "encoding/json"
  "io"
  "log"
  "net/http"


  "github.com/gorilla/mux"
  walletv1 "github.com/gusplusbus/trustflow/data_server/gen/walletv1"
  "github.com/gusplusbus/trustflow/api/internal/clients"
  "github.com/gusplusbus/trustflow/api/internal/middleware"
)


func HandlePut(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	 	uid, ok := middleware.UserIDFromCtx(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
  // Project scope (already fetched by WithProjectContext)
	pc, ok := middleware.ProjectCtx(r)
	if !ok || pc.Project == nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	projectID := mux.Vars(r)["id"]
	if projectID == "" {
		http.Error(w, "missing project id", http.StatusBadRequest)
		return
	}


  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
    return
  }
  var req UpsertRequest
  if err := json.Unmarshal(body, &req); err != nil {
    http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
    return
  }
  if req.Address == "" || req.ChainID == 0 {
    http.Error(w, "address and chainId required", http.StatusBadRequest)
    return
  }


  _, err = clients.WalletClient().Upsert(ctx, &walletv1.UpsertRequest{
    UserId: uid,
    ProjectId: projectID,
    Address: req.Address,
    ChainId: req.ChainID,
  })
  if err != nil {
    log.Printf("[wallet] Upsert: %v", err)
    http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
    return
  }
  w.WriteHeader(http.StatusNoContent)
}
