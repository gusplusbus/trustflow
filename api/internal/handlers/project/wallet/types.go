package wallet

type Response struct {
  Address string `json:"address"`
  ChainID int32 `json:"chainId"`
  UpdatedAt string `json:"updatedAt,omitempty"`
}

type UpsertRequest struct {
  Address string `json:"address"`
  ChainID int32 `json:"chainId"`
}
