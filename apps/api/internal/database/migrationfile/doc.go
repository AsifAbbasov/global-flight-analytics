// Package migrationfile defines the canonical PostgreSQL migration file-name contract.
//
// Every subsystem that interprets migration file identity must use this package
// so version and name semantics cannot diverge between execution, audit, and
// repair verification.
package migrationfile
