package ownership

type CreateReq struct {
	Organization string `json:"organization"`
	Repository   string `json:"repository"`
}

