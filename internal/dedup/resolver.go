package dedup

import (
	"context"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/google/uuid"
)

const MatchConfidence = 0.9

type Resolver struct{}

func NewResolver() *Resolver { return &Resolver{} }

func (r *Resolver) Resolve(ctx context.Context, props domain.PropertyRepository, candidate domain.Property) (*domain.Property, float64, error) {
	existing, err := props.FindMatch(ctx, candidate)
	if err != nil {
		return nil, 0, err
	}
	if existing != nil {
		return existing, MatchConfidence, nil
	}

	toCreate := candidate
	if toCreate.ID == uuid.Nil {
		toCreate.ID = uuid.New()
	}
	if err := props.Create(ctx, &toCreate); err != nil {
		return nil, 0, err
	}
	return &toCreate, 1, nil
}
