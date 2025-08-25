package postgres


import (
"context"
"embed"


"github.com/jackc/pgx/v5/pgxpool"


"github.com/gusplusbus/trustflow/data_server/internal/domain"
)


//go:embed queries/wallet_*.sql
var walletFS embed.FS


type WalletPG struct {
db *pgxpool.Pool


qGet string
qUpsert string
qDelete string
}


func NewWalletPG(db *pgxpool.Pool) (*WalletPG, error) {
qg, err := walletFS.ReadFile("queries/wallet_get.sql")
if err != nil { return nil, err }
qu, err := walletFS.ReadFile("queries/wallet_upsert.sql")
if err != nil { return nil, err }
qd, err := walletFS.ReadFile("queries/wallet_delete.sql")
if err != nil { return nil, err }
return &WalletPG{db: db, qGet: string(qg), qUpsert: string(qu), qDelete: string(qd)}, nil
}


func (w *WalletPG) Get(ctx context.Context, userID, projectID string) (*domain.Wallet, error) {
row := w.db.QueryRow(ctx, w.qGet, userID, projectID)
var out domain.Wallet
if err := row.Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt, &out.ProjectID, &out.UserID, &out.Address, &out.ChainID); err != nil {
return nil, err
}
return &out, nil
}


func (w *WalletPG) Upsert(ctx context.Context, userID, projectID, address string, chainID int32) (*domain.Wallet, error) {
row := w.db.QueryRow(ctx, w.qUpsert, projectID, userID, address, chainID)
var out domain.Wallet
if err := row.Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt, &out.ProjectID, &out.UserID, &out.Address, &out.ChainID); err != nil {
return nil, err
}
return &out, nil
}


func (w *WalletPG) Delete(ctx context.Context, userID, projectID string) (bool, error) {
ct, err := w.db.Exec(ctx, w.qDelete, userID, projectID)
if err != nil { return false, err }
return ct.RowsAffected() > 0, nil
}
