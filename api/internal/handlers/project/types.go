package project

type CreateReq struct {
  Title                string `json:"title"`
  Description          string `json:"description"`
  DurationEstimate     int    `json:"duration_estimate"`
  TeamSize             int    `json:"team_size"`
  ApplicationCloseTime string `json:"application_close_time"`
}

type UpdateReq struct {
	Title                *string `json:"title,omitempty"`
	Description          *string `json:"description,omitempty"`
	DurationEstimate     *int    `json:"duration_estimate,omitempty"`
	TeamSize             *int    `json:"team_size,omitempty"`
	ApplicationCloseTime *string `json:"application_close_time,omitempty"`
}

type ProjectResp struct {
  ID        string `json:"id"`
  CreatedAt string `json:"created_at"`
  UpdatedAt string `json:"updated_at"`
  Title     string `json:"title"`
  Description string `json:"description"`
  DurationEstimate int `json:"duration_estimate"`
  TeamSize int `json:"team_size"`
  ApplicationCloseTime string `json:"application_close_time"`
}
