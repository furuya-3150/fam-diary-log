package db

import (
	"context"
)

type TransactionManager interface {
	ExecuteTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type transactionManager struct {
	DBManager *DBManager
}

func NewTransaction(dm *DBManager) TransactionManager {
	return &transactionManager{dm}
}

func (t *transactionManager) ExecuteTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx := t.DBManager.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	ctxWithTx := t.DBManager.WithTx(ctx, tx)

	if err := fn(ctxWithTx); err != nil {
		tx.Rollback()
		return tx.Error
	}
	tx.Commit()

	return nil
}