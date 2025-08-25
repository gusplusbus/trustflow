package wallet


import (
  "log"
  "net/http"


  "github.com/gorilla/mux"
  walletv1 "github.com/gusplusbus/trustflow/data_server/gen/walletv1"
  "github.com/gusplusbus/trustflow/api/internal/clients"
  "github.com/gusplusbus/trustflow/api/internal/middleware"
)


func HandleDelete(w http.ResponseWriter, r *http.Request) {
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

  _, err := clients.WalletClient().Detach(ctx, &walletv1.DetachRequest{
    UserId: uid,
    ProjectId: projectID,
  })
  if err != nil {
    log.Printf("[wallet] Detach: %v", err)
    http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
    return
  }
  w.WriteHeader(http.StatusNoContent)
}
