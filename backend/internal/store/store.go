package store

import "context"

// Store will own PostgreSQL access for accounts, transactions, budgets, and encrypted credentials.
type Store struct{}

func New() Store { return Store{} }

func (s Store) Ping(ctx context.Context) error { return nil }
