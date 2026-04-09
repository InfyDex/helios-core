package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/infydex/helios-core/pkg/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// ErrDuplicateUser is returned when a unique constraint on users is violated.
var ErrDuplicateUser = errors.New("user already exists")

// Store performs user persistence via sqlc-generated queries.
type Store struct {
	q *db.Queries
}

// NewStore wraps sqlc queries.
func NewStore(q *db.Queries) *Store {
	return &Store{q: q}
}

// GetOrCreateByGoogle returns an existing user by google_id or creates one.
func (s *Store) GetOrCreateByGoogle(ctx context.Context, email, name, picture, googleID string) (db.User, error) {
	u, err := s.q.GetUserByGoogleID(ctx, googleID)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, err
	}
	var avatar pgtype.Text
	if picture != "" {
		avatar = pgtype.Text{String: picture, Valid: true}
	}
	u, err = s.q.CreateUser(ctx, db.CreateUserParams{
		Email:     email,
		Name:      name,
		AvatarUrl: avatar,
		GoogleID:  googleID,
	})
	if err == nil {
		return u, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return db.User{}, fmt.Errorf("%w", ErrDuplicateUser)
	}
	return db.User{}, err
}
