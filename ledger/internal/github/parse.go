package github

import (
	"encoding/json"
	"errors"
	"strings"
)

type MinimalEvent struct {
	Delivery  string
	Event     string // e.g. "issues:assigned"
	Owner     string
	Repo      string
	GHIssueID int64
	Number    int
}

type envelope struct {
	Action     string `json:"action"`
	Issue      *struct {
		ID     int64 `json:"id"`
		Number int   `json:"number"`
	} `json:"issue"`
	Repository *struct {
		Name  string `json:"name"`
		Owner *struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

func ParseIssuesEvent(delivery, ghEvent string, body []byte) (MinimalEvent, error) {
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return MinimalEvent{}, err
	}
	if env.Repository == nil || env.Repository.Owner == nil || env.Issue == nil {
		return MinimalEvent{}, errors.New("missing repo/owner/issue")
	}
	return MinimalEvent{
		Delivery:  strings.TrimSpace(delivery),
		Event:     strings.TrimSpace(ghEvent + ":" + env.Action),
		Owner:     strings.TrimSpace(env.Repository.Owner.Login),
		Repo:      strings.TrimSpace(env.Repository.Name),
		GHIssueID: env.Issue.ID,
		Number:    env.Issue.Number,
	}, nil
}
