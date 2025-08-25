package domain


import "time"


type Wallet struct {
ID string
ProjectID string
UserID string
Address string
ChainID int32
CreatedAt time.Time
UpdatedAt time.Time
}
