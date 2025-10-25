package domain

import (
	"context"
)

type ShortenerService interface {
	Create(ctx context.Context, longURL string) (URL, error)
	Get(ctx context.Context, shortCode string) (URL, error)
	Update(ctx context.Context, shortCode string, newURL string) (URL, error)
	Delete(ctx context.Context, shortCode string) error
	GetStats(ctx context.Context, shortCode string) (URL, error)
	Redirect(ctx context.Context, shortCode string) (string, error) // Returns the long URL for redirection
}