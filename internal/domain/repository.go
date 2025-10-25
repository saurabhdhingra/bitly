package domain

import (
	"context"
)

type Repository interface {
	// FindByShortCode retrieves a URL document by its short code.
	FindByShortCode(ctx context.Context, shortCode string) (URL, error)
	// FindByOriginalURL retrieves a URL document by its original long URL.
	FindByOriginalURL(ctx context.Context, originalURL string) (URL, error)
	// Save creates a new URL document.
	Save(ctx context.Context, u URL) (URL, error)
	// Update updates the long URL for an existing document.
	Update(ctx context.Context, shortCode string, newURL string) (URL, error)
	// IncrementAccessCount increments the access count for a short code.
	IncrementAccessCount(ctx context.Context, shortCode string) error
	// Delete removes a URL document by its short code.
	Delete(ctx context.Context, shortCode string) error
}