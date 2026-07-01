package repository

import (
	"context"
	"encoding/json"
)

type CreateAuditLogParams struct {
	ActorID    *string         `json:"actor_id,omitempty"`
	ActorRole  *string         `json:"actor_role,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   *string         `json:"entity_id,omitempty"`
	Details    json.RawMessage `json:"details,omitempty"`
}

const createAuditLog = `
INSERT INTO audit_log (actor_id, actor_role, action, entity_type, entity_id, details)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, actor_id, actor_role, action, entity_type, entity_id, details, created_at`

func (q *Queries) CreateAuditLog(ctx context.Context, p CreateAuditLogParams) (AuditLogEntry, error) {
	row := q.db.QueryRow(ctx, createAuditLog, p.ActorID, p.ActorRole, p.Action, p.EntityType, p.EntityID, p.Details)
	var e AuditLogEntry
	err := row.Scan(&e.ID, &e.ActorID, &e.ActorRole, &e.Action, &e.EntityType, &e.EntityID, &e.Details, &e.CreatedAt)
	return e, err
}

const listAuditLog = `
SELECT id, actor_id, actor_role, action, entity_type, entity_id, details, created_at
FROM audit_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`

func (q *Queries) ListAuditLog(ctx context.Context, limit, offset int32) ([]AuditLogEntry, error) {
	rows, err := q.db.Query(ctx, listAuditLog, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuditLogEntry
	for rows.Next() {
		var e AuditLogEntry
		if err := rows.Scan(&e.ID, &e.ActorID, &e.ActorRole, &e.Action, &e.EntityType, &e.EntityID, &e.Details, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

type PlatformStats struct {
	TotalUsers       int64 `json:"total_users"`
	TotalTournaments int64 `json:"total_tournaments"`
	ActiveTournaments int64 `json:"active_tournaments"`
	TotalRegistrations int64 `json:"total_registrations"`
	PaidRegistrations int64 `json:"paid_registrations"`
	TotalPaymentsPaid int64 `json:"total_payments_paid"`
	PendingDisputes  int64 `json:"pending_disputes"`
	PendingPayouts   int64 `json:"pending_payouts"`
	RevenueETB       int64 `json:"revenue_etb"`
}

const platformStats = `
SELECT
    (SELECT COUNT(*) FROM profiles) AS total_users,
    (SELECT COUNT(*) FROM tournaments) AS total_tournaments,
    (SELECT COUNT(*) FROM tournaments WHERE status IN ('open', 'registration_closed', 'in_progress')) AS active_tournaments,
    (SELECT COUNT(*) FROM registrations) AS total_registrations,
    (SELECT COUNT(*) FROM registrations WHERE payment_status = 'paid') AS paid_registrations,
    (SELECT COUNT(*) FROM payments WHERE status = 'paid') AS total_payments_paid,
    (SELECT COUNT(*) FROM matches WHERE dispute_status = 'pending') AS pending_disputes,
    (SELECT COUNT(*) FROM payouts WHERE status = 'pending') AS pending_payouts,
    (SELECT COALESCE(SUM(amount * platform_fee_pct / 100), 0)::bigint
        FROM payments p
        JOIN tournaments t ON t.id = p.tournament_id
        WHERE p.status = 'paid'
    ) AS revenue_etb`

func (q *Queries) GetPlatformStats(ctx context.Context) (PlatformStats, error) {
	var s PlatformStats
	err := q.db.QueryRow(ctx, platformStats).Scan(
		&s.TotalUsers, &s.TotalTournaments, &s.ActiveTournaments,
		&s.TotalRegistrations, &s.PaidRegistrations, &s.TotalPaymentsPaid,
		&s.PendingDisputes, &s.PendingPayouts, &s.RevenueETB,
	)
	return s, err
}
