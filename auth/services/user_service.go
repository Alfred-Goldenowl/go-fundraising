package services

import (
	"context"
	"errors"
	"fmt"
	"go-fundraising/auth/models"
	"go-fundraising/db"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

type UserService struct{}

func (s *UserService) NewUser(ctx context.Context, user models.User) error {
	stmt, names := qb.Insert(models.UserTable.Name).
		Columns(models.UserTable.Columns...).
		ToCql()

	q := gocqlx.Query(db.ScyllaSession.Query(stmt), names).BindStruct(user)

	return q.ExecRelease()
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	var user models.User

	stmt, names := qb.Select(models.UserTable.Name).Where(qb.Eq("username")).ToCql()
	q := gocqlx.Query(db.ScyllaSession.Query(stmt), names).BindMap(map[string]interface{}{
		"username": username,
	})

	if err := q.GetRelease(&user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *UserService) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	stmt, names := qb.Select(models.UserTable.Name).
		Columns("id").
		Where(qb.Eq("username")).
		Limit(1).
		ToCql()

	var result struct {
		ID gocql.UUID `db:"id"`
	}

	q := gocqlx.Query(db.ScyllaSession.Query(stmt), names).
		BindMap(map[string]interface{}{"username": username})

	err := q.Get(&result)

	log.Println(result)

	if err != nil {
		if err == gocql.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *UserService) GetUserByID(ctx context.Context, user_id gocql.UUID) (models.User, error) {
	var user models.User

	stmt, names := qb.Select(models.UserTable.Name).Where(qb.Eq("id")).ToCql()
	q := gocqlx.Query(db.ScyllaSession.Query(stmt), names).BindMap(map[string]interface{}{
		"id": user_id,
	})

	if err := q.GetRelease(&user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *UserService) SaveRefreshToken(ctx context.Context, token string, userID gocql.UUID, expiresAt time.Time) error {
	stmt, names := qb.Insert(models.RefreshTokenTable.Name).Columns(models.RefreshTokenTable.Columns...).ToCql()
	q := gocqlx.Query(db.ScyllaSession.Query(stmt), names).BindMap(map[string]interface{}{
		"refresh_token": token,
		"user_id":       userID,
		"expires_at":    expiresAt,
	})
	return q.ExecRelease()
}

func (s *UserService) ValidateRefreshToken(ctx context.Context, token string) (gocql.UUID, error) {
	var row struct {
		UserID    gocql.UUID `db:"user_id"`
		ExpiresAt time.Time  `db:"expires_at"`
	}

	stmt, names := qb.Select(models.RefreshTokenTable.Name).
		Columns("user_id", "expires_at").
		Where(qb.Eq("refresh_token")).
		ToCql()

	q := gocqlx.Query(db.ScyllaSession.Query(stmt), names).BindMap(map[string]interface{}{
		"refresh_token": token,
	})

	if err := q.GetRelease(&row); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return gocql.UUID{}, fmt.Errorf("refresh token not found")
		}
		return gocql.UUID{}, err
	}

	userID := row.UserID
	expiresAt := row.ExpiresAt

	if time.Now().After(expiresAt) {
		_ = s.DeleteRefreshToken(ctx, token)
		return gocql.UUID{}, fmt.Errorf("refresh token expired")
	}

	return userID, nil
}

func (s *UserService) DeleteRefreshToken(ctx context.Context, token string) error {
	stmt, names := qb.Delete(models.RefreshTokenTable.Name).
		Where(qb.Eq("refresh_token")).ToCql()

	q := gocqlx.Query(db.ScyllaSession.Query(stmt), names).BindMap(map[string]interface{}{
		"refresh_token": token,
	})

	return q.ExecRelease()
}
