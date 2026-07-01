package admin

import (
	"context"
	"encoding/json"

	"bona-backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo *repository.Queries
	db   *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{repo: repository.New(db), db: db}
}

// Log writes an audit record. Errors are swallowed so a failed audit write
// never blocks the primary action.
func (s *Service) Log(ctx context.Context, actorID, actorRole, action, entityType string, entityID *string, details map[string]any) {
	var detailsJSON json.RawMessage
	if details != nil {
		b, _ := json.Marshal(details)
		detailsJSON = b
	}
	var actorPtr, rolePtr *string
	if actorID != "" {
		actorPtr = &actorID
	}
	if actorRole != "" {
		rolePtr = &actorRole
	}
	s.repo.CreateAuditLog(ctx, repository.CreateAuditLogParams{
		ActorID:    actorPtr,
		ActorRole:  rolePtr,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    detailsJSON,
	})
}

func (s *Service) Stats(ctx context.Context) (repository.PlatformStats, error) {
	return s.repo.GetPlatformStats(ctx)
}

func (s *Service) ListAudit(ctx context.Context, limit, offset int32) ([]repository.AuditLogEntry, error) {
	return s.repo.ListAuditLog(ctx, limit, offset)
}
