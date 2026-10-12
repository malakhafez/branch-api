package branch

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// أكواد أخطاء PostgreSQL اللي بنحوّلها لأخطاء الـ module.
const (
	pgUniqueViolation   = "23505" // اسم مكرر (branches_name_key)
	pgStringTooLong     = "22001" // نص أطول من VARCHAR(100)
	pgNumericOutOfRange = "22003" // رقم أكبر من NUMERIC(10,2)
)

// الأعمدة اللي بنقرأها.
// to_char بيرجّع الوقت كنص "HH:MM"، و late_rate::text بيرجّع الرقم بدقته الكاملة
// (من غير ما يعدّي على float) عشان نحوّله لـ decimal.
const selectColumns = `id, name, type, COALESCE(address, ''),
	to_char(official_start_time, 'HH24:MI'),
	late_rate::text, is_active, created_at, updated_at`

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// بيتأكد وقت الـ build إن التنفيذ مطابق للـ interface.
var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Create(ctx context.Context, b Branch) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO branches
			(id, name, type, address, official_start_time, late_rate, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::numeric, $7, $8, $9)`,
		b.ID, b.Name, string(b.Type), b.Address, b.OfficialStartTime,
		b.LateRate.String(), b.IsActive, b.CreatedAt, b.UpdatedAt)
	return mapError(err)
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (Branch, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+selectColumns+` FROM branches WHERE id = $1`, id)
	b, err := scanBranch(row)
	if err != nil {
		return Branch{}, mapError(err)
	}
	return b, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]Branch, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+selectColumns+` FROM branches ORDER BY name`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	branches := []Branch{}
	for rows.Next() {
		b, err := scanBranch(rows)
		if err != nil {
			return nil, mapError(err)
		}
		branches = append(branches, b)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err)
	}
	return branches, nil
}

func (r *PostgresRepository) Update(ctx context.Context, b Branch) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE branches
		SET name = $2, address = $3, official_start_time = $4,
		    late_rate = $5::numeric, is_active = $6, updated_at = $7
		WHERE id = $1`,
		b.ID, b.Name, b.Address, b.OfficialStartTime,
		b.LateRate.String(), b.IsActive, b.UpdatedAt)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// scanBranch بتقرأ صف واحد وتحوّله لـ Branch.
func scanBranch(row pgx.Row) (Branch, error) {
	var b Branch
	var branchType, lateRate string
	err := row.Scan(&b.ID, &b.Name, &branchType, &b.Address, &b.OfficialStartTime,
		&lateRate, &b.IsActive, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return b, err
	}
	b.Type = BranchType(branchType)
	b.LateRate, err = decimal.NewFromString(lateRate)
	return b, err
}

// mapError بتحوّل أخطاء الداتا بيز لأخطاء الـ module بتاعنا.
// الـ Service بيرفض القيم الغلط قبل ما توصل هنا، فالحالات دي شبكة أمان بس.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return ErrNameTaken
		case pgStringTooLong:
			return ErrNameTooLong
		case pgNumericOutOfRange:
			return ErrLateRateTooLarge
		}
	}
	return err
}
