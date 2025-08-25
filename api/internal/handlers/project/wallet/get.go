package wallet

import (
  "encoding/json"
  "log"
  "net/http"
  "time"


  "github.com/gorilla/mux"
  walletv1 "github.com/gusplusbus/trustflow/data_server/gen/walletv1"
  "github.com/gusplusbus/trustflow/api/internal/clients"
  "github.com/gusplusbus/trustflow/api/internal/middleware"
)


func HandleGet(w http.ResponseWriter, r *http.Request) {
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


  resp, err := clients.WalletClient().Get(ctx, &walletv1.GetRequest{
    UserId: uid,
    ProjectId: projectID,
  })
  if err != nil {
    log.Printf("[wallet] Get: %v", err)
    http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
    return
  }
  wlt := resp.GetWallet()
  if wlt == nil || wlt.Address == "" {
    http.NotFound(w, r)
    return
  }
  payload := Response{
    Address: wlt.Address,
    ChainID: wlt.ChainId,
    UpdatedAt: wlt.UpdatedAt,
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  _ = json.NewEncoder(w).Encode(payload)
  _ = time.Now() // appease linters if unused
}
