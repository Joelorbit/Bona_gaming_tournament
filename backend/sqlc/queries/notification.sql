-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, title, message, link, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, type, title, message, link, metadata, read_at, created_at;

-- name: ListNotificationsByUser :many
SELECT id, user_id, type, title, message, link, metadata, read_at, created_at
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL;

-- name: MarkNotificationRead :one
UPDATE notifications SET read_at = NOW()
WHERE id = $1 AND user_id = $2 AND read_at IS NULL
RETURNING id, user_id, type, title, message, link, metadata, read_at, created_at;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications SET read_at = NOW() WHERE user_id = $1 AND read_at IS NULL;

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE id = $1 AND user_id = $2;
