package ownership

type CreateReq struct {
	Organization string `json:"organization"`
	Repository   string `json:"repository"`
	Provider     string `json:"provider,omitempty"`
	WebURL       string `json:"web_url,omitempty"`
}

type issueItem struct {
	ID        int64    `json:"id"`
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	State     string   `json:"state"`
	HTMLURL   string   `json:"html_url"`
	UserLogin string   `json:"user_login,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}
type rateInfo struct {
	Limit     int `json:"limit"`
	Remaining int `json:"remaining"`
	Reset     int `json:"reset"`
}
type listResp struct {
	Items []issueItem `json:"items"`
	Total int         `json:"total"`          // -1 when unknown for /issues
	Rate  *rateInfo   `json:"rate,omitempty"` // for the UI chip
}
