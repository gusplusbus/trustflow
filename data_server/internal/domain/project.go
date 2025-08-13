package domain

import (
	"errors"
	"strings"
	"time"
)

type Project struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time

	Title       string
	Description string

	DurationEstimate int32
	TeamSize         int32

	// store exactly what client sends for now (ISO-ish string); can tighten later
	ApplicationCloseTime string

	// Hydrated only when requested (include_ownerships=true). Not persisted directly by the project repo.
	Ownerships []*Ownership
}

func (p *Project) ValidateForCreate() error {
	if p.UserID == "" {
		return errors.New("user_id required")
	}
	if t := strings.TrimSpace(p.Title); t == "" || len(t) > 84 {
		return errors.New("title invalid")
	}
	if d := strings.TrimSpace(p.Description); d == "" || len(d) > 221 {
		return errors.New("description invalid")
	}
	if p.DurationEstimate < 0 {
		return errors.New("duration_estimate invalid")
	}
	if p.TeamSize <= 0 {
		return errors.New("team_size invalid")
	}
	return nil
}

func (p *Project) ValidateForUpdate() error {
	// allow partial update rules later; for now mirror create constraints
	return p.ValidateForCreate()
}
