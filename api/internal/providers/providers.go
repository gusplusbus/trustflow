package providers

import "context"

// RepoAccessVerifier checks whether the backend can access a given repo.
type RepoAccessVerifier interface {
	VerifyAccess(ctx context.Context, owner, repo string) error
}
