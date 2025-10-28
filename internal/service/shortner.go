package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"bitly/internal/domain"
)


const (
	ShortCodeLength = 6
	ShortCodeChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	MaxRetries      = 5 // Define a max number of retries to prevent infinite loops
)

// service implements the domain.ShortenerService interface.
type service struct {
	repo domain.Repository
}

// NewShortenerService creates a new shortener service instance.
func NewShortenerService(repo domain.Repository) domain.ShortenerService {
	// Initialize random seed once
	rand.Seed(time.Now().UnixNano()) 
	return &service{repo: repo}
}

// generateShortCode creates a random 6-character code without checking the database.
func (s *service) generateShortCode() string {
	b := make([]byte, ShortCodeLength)
	for i := range b {
		b[i] = ShortCodeChars[rand.Intn(len(ShortCodeChars))]
	}
	return string(b)
}

// isValidURL checks if a string is a valid URL structure.
func isValidURL(longURL string) bool {
	u, err := url.ParseRequestURI(longURL)
	return err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
}

// Create handles the creation of a new short URL.
func (s *service) Create(ctx context.Context, longURL string) (domain.URL, error) {
	if !isValidURL(longURL) {
		return domain.URL{}, errors.New(domain.ErrInvalidURL)
	}
	
	// 1. Check if URL already exists
	existingURL, err := s.repo.FindByOriginalURL(ctx, longURL)
	if err == nil {
		return existingURL, errors.New(domain.ErrConflict) // Return 409 Conflict if already shortened
	}
	
	// 2. Try to generate and save, retrying on shortCode collision
	for i := 0; i < MaxRetries; i++ {
		// Generate code (now fast, no DB lookup)
		shortCode := s.generateShortCode()
		
		newURL := domain.URL{
			ID:          fmt.Sprintf("%d", time.Now().UnixNano()), // Simple unique ID
			URL:         longURL,
			ShortCode:   shortCode,
			AccessCount: 0,
		}
		
		savedURL, err := s.repo.Save(ctx, newURL)
		
		if err == nil {
			return savedURL, nil // Success!
		}
		
		// If the error is a shortCode collision (ErrConflict from repository), retry the loop
		if errors.Is(err, errors.New(domain.ErrConflict)) {
			// Collision detected, continue loop to generate a new code
			continue 
		}
		
		// If it's any other error (DB error, timeout, etc.), return immediately
		return domain.URL{}, err
	}
	
	// If max retries reached, return an error
	return domain.URL{}, errors.New("failed to generate unique short code after multiple attempts")
}

// Get retrieves a URL document by its short code.
func (s *service) Get(ctx context.Context, shortCode string) (domain.URL, error) {
	return s.repo.FindByShortCode(ctx, shortCode)
}

// Update handles updating the long URL for an existing short code.
func (s *service) Update(ctx context.Context, shortCode string, newURL string) (domain.URL, error) {
	if !isValidURL(newURL) {
		return domain.URL{}, errors.New(domain.ErrInvalidURL)
	}
	return s.repo.Update(ctx, shortCode, newURL)
}

// Delete handles deleting a short URL.
func (s *service) Delete(ctx context.Context, shortCode string) error {
	return s.repo.Delete(ctx, shortCode)
}

// GetStats retrieves the URL document for statistics.
func (s *service) GetStats(ctx context.Context, shortCode string) (domain.URL, error) {
	return s.repo.FindByShortCode(ctx, shortCode)
}

// Redirect handles the redirection and access count increment.
func (s *service) Redirect(ctx context.Context, shortCode string) (string, error) {
	u, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", err // Propagate ErrNotFound
	}
	
	// Increment access count asynchronously (optional, but better for performance)
	// We run this in a separate routine to avoid blocking the redirect response.
	go func() {
		// Use a short, dedicated context for the background update
		updateCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		
		err := s.repo.IncrementAccessCount(updateCtx, shortCode)
		if err != nil {
			// In a real application, log this error
			fmt.Printf("Error incrementing access count for %s: %v\n", shortCode, err)
		}
	}()
	
	return u.URL, nil
}
