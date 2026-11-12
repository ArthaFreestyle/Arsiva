package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ArthaFreestyle/Arsiva/internal/model"
)

type NotificationRepository struct{ DB *pgxpool.Pool }

const notificationColumns = `notification_id::text,type,title,message,payload,created_at,read_at`

func scanNotification(row pgx.Row) (model.Notification, error) {
	var n model.Notification
	err := row.Scan(&n.NotificationID, &n.Type, &n.Title, &n.Message, &n.Payload, &n.CreatedAt, &n.ReadAt)
	return n, err
}
func (r *NotificationRepository) List(ctx context.Context, member, page, size int, unread bool) ([]model.Notification, int, error) {
	result := []model.Notification{}
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback(ctx)
	var total int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM member_notifications WHERE member_id=$1 AND (NOT $2 OR read_at IS NULL)`, member, unread).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := tx.Query(ctx, `SELECT `+notificationColumns+` FROM member_notifications WHERE member_id=$1 AND (NOT $2 OR read_at IS NULL) ORDER BY created_at DESC,notification_id DESC LIMIT $3 OFFSET $4`, member, unread, size, (page-1)*size)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, n)
	}
	return result, total, rows.Err()
}
func (r *NotificationRepository) Count(ctx context.Context, member int) (int, error) {
	var n int
	err := r.DB.QueryRow(ctx, `SELECT count(*) FROM member_notifications WHERE member_id=$1 AND read_at IS NULL`, member).Scan(&n)
	return n, err
}
func (r *NotificationRepository) Read(ctx context.Context, member int, id int64) (model.Notification, error) {
	return scanNotification(r.DB.QueryRow(ctx, `UPDATE member_notifications SET read_at=COALESCE(read_at,NOW()) WHERE member_id=$1 AND notification_id=$2 RETURNING `+notificationColumns, member, id))
}
