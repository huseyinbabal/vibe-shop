// Package migrations embeds the SQL migration files so the server can apply
// them at startup without the files needing to be present on disk next to the
// binary.
package migrations

import "embed"

// FS holds every .sql migration. They are applied in filename order, so the
// numeric prefix (0001_, 0002_, …) defines the sequence.
//
//go:embed *.sql
var FS embed.FS
