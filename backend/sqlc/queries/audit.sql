-- name: CreateAuditLog :one
INSERT INTO audit_log (actor_id, actor_role, action, entity_type, entity_id, details)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, actor_id, actor_role, action, entity_type, entity_id, details, created_at;

-- name: ListAuditLog :many
SELECT id, actor_id, actor_role, action, entity_type, entity_id, details, created_at
FROM audit_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAuditLogByEntity :many
SELECT id, actor_id, actor_role, action, entity_type, entity_id, details, created_at
FROM audit_log
WHERE entity_type = $1 AND entity_id = $2
ORDER BY created_at DESC;
