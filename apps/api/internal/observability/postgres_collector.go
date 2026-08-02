package observability

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCollector struct {
	pool *pgxpool.Pool
}

func NewPostgresCollector(
	pool *pgxpool.Pool,
) (*PostgresCollector, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres metrics pool is required")
	}
	return &PostgresCollector{pool: pool}, nil
}

func (
	collector *PostgresCollector,
) Name() string {
	return "postgres"
}

func (
	collector *PostgresCollector,
) WritePrometheus(
	ctx context.Context,
	builder *strings.Builder,
) error {
	if collector == nil || collector.pool == nil {
		return fmt.Errorf("postgres metrics collector is unavailable")
	}
	if ctx == nil {
		return fmt.Errorf("postgres metrics context is required")
	}

	collectionContext, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	writePostgresPoolMetrics(builder, collector.pool.Stat())
	if err := collector.writeIngestionState(collectionContext, builder); err != nil {
		return fmt.Errorf("collect ingestion state: %w", err)
	}
	if err := collector.writeReconciliationState(collectionContext, builder); err != nil {
		return fmt.Errorf("collect reconciliation state: %w", err)
	}
	return nil
}

func writePostgresPoolMetrics(
	builder *strings.Builder,
	stat *pgxpool.Stat,
) {
	gauges := []struct {
		name  string
		help  string
		value int64
	}{
		{
			name:  "postgres_pool_acquired_connections",
			help:  "Current number of acquired PostgreSQL pool connections.",
			value: int64(stat.AcquiredConns()),
		},
		{
			name:  "postgres_pool_idle_connections",
			help:  "Current number of idle PostgreSQL pool connections.",
			value: int64(stat.IdleConns()),
		},
		{
			name:  "postgres_pool_total_connections",
			help:  "Current total number of PostgreSQL pool connections.",
			value: int64(stat.TotalConns()),
		},
		{
			name:  "postgres_pool_max_connections",
			help:  "Configured maximum number of PostgreSQL pool connections.",
			value: int64(stat.MaxConns()),
		},
	}
	for _, gauge := range gauges {
		metricName := metricNamespace + "_" + gauge.name
		writeHelpAndType(builder, metricName, gauge.help, "gauge")
		fmt.Fprintf(builder, "%s %d\n", metricName, gauge.value)
	}

	emptyAcquireMetric := metricNamespace + "_postgres_pool_empty_acquire_total"
	writeHelpAndType(
		builder,
		emptyAcquireMetric,
		"Total PostgreSQL acquisitions that waited for an available connection.",
		"counter",
	)
	fmt.Fprintf(builder, "%s %d\n", emptyAcquireMetric, stat.EmptyAcquireCount())
}

func (
	collector *PostgresCollector,
) writeIngestionState(
	ctx context.Context,
	builder *strings.Builder,
) error {
	rows, err := collector.pool.Query(
		ctx,
		`SELECT status, count(*)::bigint
		 FROM ingestion_runs
		 GROUP BY status`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	counts := map[string]int64{
		"running": 0,
		"success": 0,
		"failed":  0,
		"partial": 0,
	}
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return err
		}
		if _, supported := counts[status]; supported {
			counts[status] = count
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	metricName := metricNamespace + "_ingestion_runs"
	writeHelpAndType(
		builder,
		metricName,
		"Persisted ingestion runs partitioned by bounded lifecycle status.",
		"gauge",
	)
	statuses := []string{"failed", "partial", "running", "success"}
	for _, status := range statuses {
		fmt.Fprintf(
			builder,
			"%s%s %d\n",
			metricName,
			formatLabels([]label{{name: "status", value: status}}),
			counts[status],
		)
	}

	var latestFinishedAge float64
	if err := collector.pool.QueryRow(
		ctx,
		`SELECT COALESCE(
			EXTRACT(EPOCH FROM (now() - MAX(finished_at)))::double precision,
			-1
		 )
		 FROM ingestion_runs
		 WHERE finished_at IS NOT NULL`,
	).Scan(&latestFinishedAge); err != nil {
		return err
	}
	writeHelpAndType(
		builder,
		metricNamespace+"_ingestion_latest_finished_age_seconds",
		"Age in seconds of the latest finished ingestion run, or -1 when no run has finished.",
		"gauge",
	)
	fmt.Fprintf(
		builder,
		"%s_ingestion_latest_finished_age_seconds %s\n",
		metricNamespace,
		formatFloat(latestFinishedAge),
	)

	var oldestRunningAge float64
	if err := collector.pool.QueryRow(
		ctx,
		`SELECT COALESCE(
			EXTRACT(EPOCH FROM (now() - MIN(started_at)))::double precision,
			0
		 )
		 FROM ingestion_runs
		 WHERE status = 'running'`,
	).Scan(&oldestRunningAge); err != nil {
		return err
	}
	writeHelpAndType(
		builder,
		metricNamespace+"_ingestion_oldest_running_age_seconds",
		"Age in seconds of the oldest running ingestion run.",
		"gauge",
	)
	fmt.Fprintf(
		builder,
		"%s_ingestion_oldest_running_age_seconds %s\n",
		metricNamespace,
		formatFloat(oldestRunningAge),
	)
	return nil
}

func (
	collector *PostgresCollector,
) writeReconciliationState(
	ctx context.Context,
	builder *strings.Builder,
) error {
	rows, err := collector.pool.Query(
		ctx,
		`SELECT status, count(*)::bigint
		 FROM derived_reconciliation_tasks
		 GROUP BY status`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	counts := map[string]int64{
		"pending":    0,
		"processing": 0,
		"completed":  0,
		"failed":     0,
	}
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return err
		}
		if _, supported := counts[status]; supported {
			counts[status] = count
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	metricName := metricNamespace + "_reconciliation_tasks"
	writeHelpAndType(
		builder,
		metricName,
		"Persisted reconciliation tasks partitioned by bounded lifecycle status.",
		"gauge",
	)
	statuses := make([]string, 0, len(counts))
	for status := range counts {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)
	for _, status := range statuses {
		fmt.Fprintf(
			builder,
			"%s%s %d\n",
			metricName,
			formatLabels([]label{{name: "status", value: status}}),
			counts[status],
		)
	}

	var oldestPendingAge float64
	if err := collector.pool.QueryRow(
		ctx,
		`SELECT COALESCE(
			EXTRACT(EPOCH FROM (now() - MIN(created_at)))::double precision,
			0
		 )
		 FROM derived_reconciliation_tasks
		 WHERE status = 'pending'`,
	).Scan(&oldestPendingAge); err != nil {
		return err
	}
	writeHelpAndType(
		builder,
		metricNamespace+"_reconciliation_oldest_pending_age_seconds",
		"Age in seconds of the oldest pending reconciliation task.",
		"gauge",
	)
	fmt.Fprintf(
		builder,
		"%s_reconciliation_oldest_pending_age_seconds %s\n",
		metricNamespace,
		formatFloat(oldestPendingAge),
	)
	return nil
}
