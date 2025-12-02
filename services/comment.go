package services

import (
	"context"
	"time"

	"github.com/gocql/gocql"

	"go-fundraising/db"
	"go-fundraising/models"

	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

type CommentService struct{}

func (s *CommentService) InsertComment(ctx context.Context, comment models.Comment) error {
	stmt, names := qb.Insert(models.CommentTable.Name).
		Columns(models.CommentTable.Columns...).
		ToCql()

	return gocqlx.Query(db.ScyllaSession.Query(stmt).Consistency(gocql.One), names).
		BindStruct(comment).
		ExecRelease()
}

func (s *CommentService) GetCommentsByPostID(ctx context.Context, postID string, perPage int, lastCreatedAt time.Time) ([]models.Comment, error) {

	qbSelect := qb.Select(models.CommentTable.Name).
		Where(qb.Eq("post_id")).
		OrderBy("created_at", qb.DESC).
		Limit(uint(perPage))

	if !lastCreatedAt.IsZero() {
		qbSelect = qbSelect.Where(qb.Lt("created_at"))
	}

	stmt, names := qbSelect.ToCql()

	var comments []models.Comment
	bindParams := qb.M{"post_id": postID}

	if !lastCreatedAt.IsZero() {
		bindParams["created_at"] = lastCreatedAt
	}

	q := gocqlx.Query(db.ScyllaSession.Query(stmt).Consistency(gocql.One), names).
		BindMap(bindParams)

	if err := q.SelectRelease(&comments); err != nil {
		return nil, err
	}

	return comments, nil
}
