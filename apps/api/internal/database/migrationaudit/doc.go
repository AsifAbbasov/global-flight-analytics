// Package migrationaudit performs read-only reconciliation between local
// SQL migration files and the database schema_migrations history.
//
// The package never creates the schema_migrations table, executes migration
// SQL, or changes applied history. Its only database operations are SELECT
// statements.
package migrationaudit
