package core_worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	core_postgres_pool "github.com/DimaKirejko/Dstributed_cron/internal/core/repository/postgres_pgx"
	"github.com/jackc/pgx/v5"
)

type DBActionRepository struct {
	pool core_postgres_pool.PgxPool
}

type tableRef struct {
	Schema string
	Table  string
}

func NewDBActionRepository(pool core_postgres_pool.PgxPool) *DBActionRepository {
	return &DBActionRepository{pool: pool}
}

func (r *DBActionRepository) CreatePartition(ctx context.Context, targetDB string) error {
	target, err := parseTableRef(targetDB)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFatalDBAction, err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	runDate, err := r.currentDate(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to get table time %w", ErrRetryableDBAction, err)
	}

	parentTable, err := r.rangePartitionedTable(ctx, target)
	if err != nil {
		return fmt.Errorf("%w: failed to find table %w", ErrFatalDBAction, err)
	}

	for partitionedDateRange := 0; partitionedDateRange < 7; partitionedDateRange++ {
		partitionDate := runDate.AddDate(0, 0, partitionedDateRange)

		if err := r.createDailyPartition(ctx, parentTable.Schema, parentTable.Table, partitionDate); err != nil {
			return fmt.Errorf("%w: create daily partition: %w", ErrRetryableDBAction, err)
		}
	}

	return nil
}

func (r *DBActionRepository) YearCleanup(ctx context.Context, targetDB string) error {
	target, err := parseTableRef(targetDB)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFatalDBAction, err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	if err := r.ensureCleanupTarget(ctx, target); err != nil {
		return fmt.Errorf("%w: cleanup target check failed: %w", ErrFatalDBAction, err)
	}

	if err := r.deleteRowsOlderThanYear(ctx, target); err != nil {
		return fmt.Errorf("%w: delete rows older than one year: %w", ErrRetryableDBAction, err)
	}

	return nil
}

func (r *DBActionRepository) currentDate(ctx context.Context) (time.Time, error) {
	var currentDate time.Time

	if err := r.pool.QueryRow(ctx, `SELECT current_date;`).Scan(&currentDate); err != nil {
		return time.Time{}, fmt.Errorf("get current date: %w", err)
	}

	return currentDate, nil
}

func (r *DBActionRepository) rangePartitionedTable(ctx context.Context, target tableRef) (tableRef, error) {
	query := `
	SELECT n.nspname, c.relname
	FROM pg_class c
	JOIN pg_namespace n ON n.oid = c.relnamespace
	JOIN pg_partitioned_table pt ON pt.partrelid = c.oid
	WHERE n.nspname = $1
  	AND c.relname = $2
  	AND pt.partstrat = 'r';`

	rows, err := r.pool.Query(ctx, query, target.Schema, target.Table)
	if err != nil {
		return tableRef{}, fmt.Errorf("find range partitioned table: %w", err)
	}
	defer rows.Close()

	var tables []tableRef
	for rows.Next() {
		var table tableRef
		if err := rows.Scan(&table.Schema, &table.Table); err != nil {
			return tableRef{}, fmt.Errorf("scan partitioned table: %w", err)
		}

		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return tableRef{}, fmt.Errorf("iterate partitioned tables: %w", err)
	}

	if len(tables) == 0 {
		return tableRef{}, fmt.Errorf("range partitioned table not found: %s.%s", target.Schema, target.Table)
	}

	if len(tables) > 1 {
		return tableRef{}, fmt.Errorf("range partitioned table is ambiguous: %s.%s", target.Schema, target.Table)
	}

	return tables[0], nil
}

func (r *DBActionRepository) ensureCleanupTarget(ctx context.Context, target tableRef) error {
	query := `
	SELECT EXISTS (
		SELECT 1
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_attribute a ON a.attrelid = c.oid
		WHERE n.nspname = $1
			AND c.relname = $2
			AND c.relkind IN ('r', 'p')
			AND a.attname = 'created_at'
			AND NOT a.attisdropped
	);`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, target.Schema, target.Table).Scan(&exists); err != nil {
		return fmt.Errorf("check cleanup target: %w", err)
	}

	if !exists {
		return fmt.Errorf("cleanup target must exist and contain created_at column: %s.%s", target.Schema, target.Table)
	}

	return nil
}

func (r *DBActionRepository) deleteRowsOlderThanYear(ctx context.Context, target tableRef) error {
	query := fmt.Sprintf(
		`DELETE FROM %s WHERE created_at < current_date - INTERVAL '1 year';`,
		pgx.Identifier{target.Schema, target.Table}.Sanitize(),
	)

	if _, err := r.pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("delete from %s.%s: %w", target.Schema, target.Table, err)
	}

	return nil
}

func (r *DBActionRepository) createDailyPartition(
	ctx context.Context,
	schema string,
	parentTable string,
	runDate time.Time,
) error {
	from := runDate.Format("2006-01-02")
	to := runDate.AddDate(0, 0, 1).Format("2006-01-02")
	partitionTable := dailyPartitionName(parentTable, runDate)

	query := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s');`,
		pgx.Identifier{schema, partitionTable}.Sanitize(),
		pgx.Identifier{schema, parentTable}.Sanitize(),
		from,
		to,
	)

	if _, err := r.pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("create daily partition %s.%s: %w", schema, partitionTable, err)
	}

	return nil
}

func dailyPartitionName(parentTable string, runDate time.Time) string {
	const maxIdentifierLength = 63

	suffix := "_" + runDate.Format("20060102")
	maxParentLength := maxIdentifierLength - len(suffix)

	if len(parentTable) > maxParentLength {
		parentTable = parentTable[:maxParentLength]
	}

	return parentTable + suffix
}

func parseTableRef(input string) (tableRef, error) {
	parts := strings.Split(input, ".")

	if len(parts) != 2 {
		return tableRef{}, fmt.Errorf("target table must be in schema.table format: %q", input)
	}

	return tableRef{
		Schema: parts[0],
		Table:  parts[1],
	}, nil
}
