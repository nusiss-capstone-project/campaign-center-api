package mysql

import (
	"sync"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

// AuditRepository appends audit rows.
type AuditRepository interface {
	Create(a *model.AuditLog) error
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

func (r *auditRepository) Create(a *model.AuditLog) error {
	if DB == nil {
		return ErrDatabaseDisabled
	}
	return DB.Create(a).Error
}
