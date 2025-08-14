package domain

import "time"

// Minimal fields to render back to the UI and keep uniqueness.
// We keep org/repo/number so uniqueness is per project+repo+number.
type Issue struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time

	ProjectID string
	UserID    string

	Organization string
	Repository   string

	GHIssueID   int64  // GH internal id
	GHNumber    int32  // issue number
	Title       string
	State       string
	HTMLURL     string
	Labels      []string
	GHUserLogin string

	GHCreatedAt time.Time
	GHUpdatedAt time.Time
}
