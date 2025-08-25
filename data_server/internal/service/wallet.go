package service


import (
"context"
"errors"


"github.com/gusplusbus/trustflow/data_server/internal/domain"
"github.com/gusplusbus/trustflow/data_server/internal/repo"
)


type WalletService struct {
repo repo.WalletRepo
}


func NewWalletService(r repo.WalletRepo) *WalletService { return &WalletService{repo: r} }


func (s *WalletService) Get(ctx context.Context, userID, projectID string) (*domain.Wallet, error) {
return s.repo.Get(ctx, userID, projectID)
}


func (s *WalletService) Upsert(ctx context.Context, userID, projectID, address string, chainID int32) (*domain.Wallet, error) {
if address == "" { return nil, errors.New("address required") }
if chainID == 0 { return nil, errors.New("chain_id required") }
return s.repo.Upsert(ctx, userID, projectID, address, chainID)
}


func (s *WalletService) Detach(ctx context.Context, userID, projectID string) (bool, error) {
return s.repo.Delete(ctx, userID, projectID)
}
