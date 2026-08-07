package mysql

import (
	"context"
	"sync"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

// AuditRepository appends audit rows.
type AuditRepository interface {
	Create(ctx context.Context, a *model.AuditLog) error
}

type auditRepository struct{}

var (
	auditRepositoryOnce     sync.Once
	auditRepositoryInstance AuditRepository
)

// GetAuditRepository returns the singleton audit repository.
func GetAuditRepository() AuditRepository {
	auditRepositoryOnce.Do(func() {
		auditRepositoryInstance = &auditRepository{}
	})
	return auditRepositoryInstance
}

func (r *auditRepository) Create(ctx context.Context, a *model.AuditLog) error {
	db, err := session(ctx)
	if err != nil {
		return err
	}
	return db.Create(a).Error
}
