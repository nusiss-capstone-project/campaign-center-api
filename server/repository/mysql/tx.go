package mysql

import (
	"context"

	"gorm.io/gorm"
)

type txContextKey struct{}

// WithinTransaction runs fn inside a DB transaction. The transactional *gorm.DB
// is stored on the context so repositories can participate without accepting a
// *gorm.DB parameter. When DB is nil (e.g. unit tests with in-memory repos),
// fn is invoked with the original context and no transaction is started.
func WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if DB == nil {
		return fn(ctx)
	}
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txContextKey{}, tx))
	})
}

// session returns the request-scoped DB handle: an active transaction if present,
// otherwise the global DB. Always attaches ctx for tracing/logging.
func session(ctx context.Context) (*gorm.DB, error) {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx), nil
	}
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB.WithContext(ctx), nil
}
