// Package migrationrepair verifies that the deployed PostgreSQL schema and
// migration history are safe for the repository migration sequence repair.
//
// The package is read-only. It never updates schema_migrations and never
// creates, alters, or drops database objects.
package migrationrepair
