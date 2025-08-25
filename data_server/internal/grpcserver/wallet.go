package grpcserver

import (
	"context"
	"time"

	walletv1 "github.com/gusplusbus/trustflow/data_server/gen/walletv1"
	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/service"
)
  
type WalletServer struct {
walletv1.UnimplementedWalletServiceServer
svc *service.WalletService
}


func NewWalletServer(s *service.WalletService) *WalletServer { return &WalletServer{svc: s} }


func (s *WalletServer) Get(ctx context.Context, in *walletv1.GetRequest) (*walletv1.GetResponse, error) {
w, err := s.svc.Get(ctx, in.GetUserId(), in.GetProjectId())
if err != nil { return &walletv1.GetResponse{}, nil }
if w == nil { return &walletv1.GetResponse{}, nil }
return &walletv1.GetResponse{Wallet: toPb(w)}, nil
}


func (s *WalletServer) Upsert(ctx context.Context, in *walletv1.UpsertRequest) (*walletv1.UpsertResponse, error) {
w, err := s.svc.Upsert(ctx, in.GetUserId(), in.GetProjectId(), in.GetAddress(), in.GetChainId())
if err != nil { return nil, err }
return &walletv1.UpsertResponse{Wallet: toPb(w)}, nil
}


func (s *WalletServer) Detach(ctx context.Context, in *walletv1.DetachRequest) (*walletv1.DetachResponse, error) {
ok, err := s.svc.Detach(ctx, in.GetUserId(), in.GetProjectId())
if err != nil { return nil, err }
return &walletv1.DetachResponse{Ok: ok}, nil
}


func toPb(w *domain.Wallet) *walletv1.Wallet {
return &walletv1.Wallet{
Id: w.ID,
ProjectId: w.ProjectID,
UserId: w.UserID,
Address: w.Address,
ChainId: w.ChainID,
CreatedAt: w.CreatedAt.UTC().Format(time.RFC3339),
UpdatedAt: w.UpdatedAt.UTC().Format(time.RFC3339),
}
}
