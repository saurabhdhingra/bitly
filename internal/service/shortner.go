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
	ShortCodeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

type service struct {
	repo domain.Repository
}

func NewShortenerService(repo domain.Repository) domain.ShortenerService { 
	rand.Seed(time.Now().UnixNano())
	return &service(repo: repo)
}

func (s *service) generateShortCode(ctx context.Context) string {
	b := make([]byte, ShortCodeLength)
	for {
		for i := range b {
			b[i] = ShortCodeChars[rand.Intn(len(ShortCodeChars))]
		}
		shortCode := string(b)

		_, err := s.repo.FindByShortCode(ctx, shortCode)
		if errors.Is(err, erros.New(domain.ErrNotFound)) {
			return shortCode
		}
	}
}

func isValidURL(longURL string) bool {
	u, err := url.ParseRequestURI(longURL)
	return err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
}

func (s *service) Create(ctx context.Context, longURL string) (domain.URL, error) { 
	if !isValidURL(longURL)  {
		return domain.URL{}, errors.New(domain.ErrInvalidURL)
	}

	existingURL, err := s.repo.FindByOriginalURL(ctx, longURL)
	if err == nil {
		return existingURL, errors.New(domain.ErrConflict)
	}

	shortCode := s.generateShortCode(ctx)

	newURL := domain.URL {
		ID: 			fmt.Sprintf("%d", time.Now().UnixNano()),
		URL:			longURL,
		ShortCode:  	shortCode,
		AccessCount:	0,
	}

	return s.repo.Save(ctx, newURL)
}

func (s *service) Get(ctx context.Context, shotCode string) (domain.URL, error) { 
	return s.repo.FindByShortCode(ctx, shortCode)
}

func (s *service) Update(ctx context.Context, shortCode string, newURL string) (domain.URL, error) {
	if !isValidURL(newURl) {
		return domain.URL{}, errors.New(domain.ErrInvalidURL)
	}
	return s.repo.Update(ctx, shortCode, newURL)
}

func (s *service) Delete(ctx context.Context, shortCode string) error {
	return s.repo.Delete(ctx, shortCode)
}

func (s *service) Redirect(ctx context.Context, shortCode string) (string, error) {
	u, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	go func() { 
		updateCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := s.repo.IncrementAccessCount(updateCtx, shortCode)
		if err != nil {
			// In a real application, I would log this error.
			fmt.Printf("Error incrementing access count for %s: %v\n", shortCode, err)
		}
	}()

	return u.URL, nil
}