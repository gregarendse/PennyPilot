package migrations

import "embed"

// FS contains the database migration files used by the application at runtime.
//
//go:embed *.sql sqlite/*.sql
var FS embed.FS
