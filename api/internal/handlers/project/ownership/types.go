package ownership

type CreateReq struct {
	Organization string `json:"organization"`
	Repository   string `json:"repository"`
	Provider     string `json:"provider,omitempty"`
	WebURL       string `json:"web_url,omitempty"`
}

