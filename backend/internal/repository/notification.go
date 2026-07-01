package repository

import (
	"context"
	"encoding/json"
)

type CreateNotificationParams struct {
	UserID   string          `json:"user_id"`
	Type     string          `json:"type"`
	Title    string          `json:"title"`
	Message  string          `json:"message"`
	Link     *string         `json:"link,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

const createNotification = `
INSERT INTO notifications (user_id, type, title, message, link, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, type, title, message, link, metadata, read_at, created_at`

func (q *Queries) CreateNotification(ctx context.Context, p CreateNotificationParams) (Notification, error) {
	row := q.db.QueryRow(ctx, createNotification, p.UserID, p.Type, p.Title, p.Message, p.Link, p.Metadata)
	var n Notification
	err := row.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.Link, &n.Metadata, &n.ReadAt, &n.CreatedAt)
	return n, err
}

const listNotificationsByUser = `
SELECT id, user_id, type, title, message, link, metadata, read_at, created_at
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

func (q *Queries) ListNotificationsByUser(ctx context.Context, userID string, limit, offset int32) ([]Notification, error) {
	rows, err := q.db.Query(ctx, listNotificationsByUser, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.Link, &n.Metadata, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

const countUnreadNotifications = `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`

func (q *Queries) CountUnreadNotifications(ctx context.Context, userID string) (int64, error) {
	row := q.db.QueryRow(ctx, countUnreadNotifications, userID)
	var c int64
	err := row.Scan(&c)
	return c, err
}

const markNotificationRead = `
UPDATE notifications SET read_at = NOW()
WHERE id = $1 AND user_id = $2 AND read_at IS NULL
RETURNING id, user_id, type, title, message, link, metadata, read_at, created_at`

func (q *Queries) MarkNotificationRead(ctx context.Context, id, userID string) (Notification, error) {
	row := q.db.QueryRow(ctx, markNotificationRead, id, userID)
	var n Notification
	err := row.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.Link, &n.Metadata, &n.ReadAt, &n.CreatedAt)
	return n, err
}

const markAllNotificationsRead = `UPDATE notifications SET read_at = NOW() WHERE user_id = $1 AND read_at IS NULL`

func (q *Queries) MarkAllNotificationsRead(ctx context.Context, userID string) error {
	_, err := q.db.Exec(ctx, markAllNotificationsRead, userID)
	return err
}

const deleteNotification = `DELETE FROM notifications WHERE id = $1 AND user_id = $2`

func (q *Queries) DeleteNotification(ctx context.Context, id, userID string) error {
	_, err := q.db.Exec(ctx, deleteNotification, id, userID)
	return err
}
