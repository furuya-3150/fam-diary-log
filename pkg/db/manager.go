package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	txKey = "tx"
)

// DBManager manages database connections and transactions
type DBManager struct {
	db *gorm.DB
}

// NewDBManager creates a new DBManager instance
func NewDBManager(dbUrl string) *DBManager {
	if dbUrl == "" {
		slog.Error("database URL is required")
		os.Exit(1)
	}
	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect db", err)
		os.Exit(1)
	}

	// Pingで接続確認
	if _, err := db.DB(); err != nil {
		slog.Error("DB ping failed: %v", err)
		os.Exit(1)
	}

	slog.Info("DB Connceted")
	return &DBManager{db}
}

// DB returns the database instance with context
func (dm *DBManager) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(txKey).(*gorm.DB)
	if !ok {
		return dm.db.WithContext(ctx)
	}

	return tx.WithContext(ctx)
}

// TransactionManager defines the interface for transaction management
type TransactionManager interface {
	BeginTx(ctx context.Context) (context.Context, error)
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error
}

// transactionManagerImpl implements TransactionManager
type transactionManagerImpl struct {
	manager *DBManager
}

// NewTransaction creates a new TransactionManager instance
func NewTransaction(manager *DBManager) TransactionManager {
	return &transactionManagerImpl{manager: manager}
}

// BeginTx begins a transaction and stores it in the context
func (t *transactionManagerImpl) BeginTx(ctx context.Context) (context.Context, error) {
	tx := t.manager.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	return context.WithValue(ctx, txKey, tx), nil
}

// CommitTx commits the transaction
func (t *transactionManagerImpl) CommitTx(ctx context.Context) error {
	tx, ok := ctx.Value(txKey).(*gorm.DB)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}

	result := tx.Commit()
	if result.Error != nil {
		return fmt.Errorf("failed to commit transaction: %w", result.Error)
	}

	return nil
}

// RollbackTx rolls back the transaction
func (t *transactionManagerImpl) RollbackTx(ctx context.Context) error {
	tx, ok := ctx.Value(txKey).(*gorm.DB)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}

	result := tx.Rollback()
	if result.Error != nil {
		return fmt.Errorf("failed to rollback transaction: %w", result.Error)
	}

	return nil
}

func (dm *DBManager) GetGorm() *gorm.DB {
	return dm.db
}
