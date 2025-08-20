package timeline

import "time"

type Item struct {
	Provider        string                 `json:"provider"`           // "github"
	ProviderEventID string                 `json:"provider_event_id"`  // GraphQL node id
	IssueNodeID     string                 `json:"issue_node_id"`      // GraphQL node id of the issue (when available)
	Type            string                 `json:"type"`               // e.g., LabeledEvent, IssueComment, AssignedEvent
	Actor           string                 `json:"actor,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	Payload         map[string]any         `json:"payload"`            // raw-ish fields for the typename
}
