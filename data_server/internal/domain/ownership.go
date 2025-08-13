package domain

import (
	"errors"
	"strings"
	"time"
)

type Ownership struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time

	ProjectID    string
	UserID       string
	Organization string
	Repository   string
	Provider     string
	WebURL       string
}

func (o *Ownership) ValidateForCreate() error {
	if o.UserID == "" {
		return errors.New("user_id required")
	}
	if o.ProjectID == "" {
		return errors.New("project_id required")
	}
	if s := strings.TrimSpace(o.Organization); s == "" {
		return errors.New("organization required")
	}
	if s := strings.TrimSpace(o.Repository); s == "" {
		return errors.New("repository required")
	}
	return nil
}
