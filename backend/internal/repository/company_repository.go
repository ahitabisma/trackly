package repository

import (
	"context"
	"fmt"

	"trackly-backend/internal/model"
	"trackly-backend/pkg/filter"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyRepository interface {
	FindAll(ctx context.Context, fq filter.FilteringQuery) ([]model.Company, int64, error)
	FindByID(ctx context.Context, id int) (*model.Company, error)
	Create(ctx context.Context, company *model.Company) (*model.Company, error)
	Update(ctx context.Context, id int, company *model.Company) (*model.Company, error)
	Delete(ctx context.Context, id int) error
	BulkCreate(ctx context.Context, companies []model.Company) ([]model.Company, []error, error)
}

type companyRepository struct {
	db *pgxpool.Pool
}

func NewCompanyRepository(db *pgxpool.Pool) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) FindAll(ctx context.Context, fq filter.FilteringQuery) ([]model.Company, int64, error) {

	allowed := []string{"kode", "nama_perusahaan"}

	qr := filter.BuildSQL(fq, allowed, "kode ASC")

	base := "FROM companies"

	// COUNT
	var total int64
	countQuery := "SELECT COUNT(*) " + base + qr.Where
	err := r.db.QueryRow(ctx, countQuery, qr.Args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// DATA
	query := fmt.Sprintf(`
		SELECT id, kode, nama_perusahaan, papan_pencatatan
		%s %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, base, qr.Where, qr.Order, len(qr.Args)+1, len(qr.Args)+2)

	args := append(qr.Args, qr.Limit, qr.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []model.Company

	for rows.Next() {
		var c model.Company
		err := rows.Scan(&c.ID, &c.Kode, &c.NamaPerusahaan, &c.PapanPencatatan)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, c)
	}

	return result, total, nil
}

func (r *companyRepository) FindByID(ctx context.Context, id int) (*model.Company, error) {
	query := `
		SELECT id, kode, nama_perusahaan, tanggal_pencatatan, jumlah_saham, papan_pencatatan, created_at, updated_at
		FROM companies
		WHERE id = $1
	`

	var c model.Company
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Kode, &c.NamaPerusahaan, &c.TanggalPencatatan,
		&c.JumlahSaham, &c.PapanPencatatan, &c.CreatedAt, &c.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *companyRepository) Create(ctx context.Context, company *model.Company) (*model.Company, error) {
	query := `
		INSERT INTO companies (kode, nama_perusahaan, tanggal_pencatatan, jumlah_saham, papan_pencatatan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, kode, nama_perusahaan, tanggal_pencatatan, jumlah_saham, papan_pencatatan, created_at, updated_at
	`

	var c model.Company
	err := r.db.QueryRow(ctx, query,
		company.Kode,
		company.NamaPerusahaan,
		company.TanggalPencatatan,
		company.JumlahSaham,
		company.PapanPencatatan,
	).Scan(
		&c.ID, &c.Kode, &c.NamaPerusahaan, &c.TanggalPencatatan,
		&c.JumlahSaham, &c.PapanPencatatan, &c.CreatedAt, &c.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *companyRepository) Update(ctx context.Context, id int, company *model.Company) (*model.Company, error) {
	query := `
		UPDATE companies
		SET kode = COALESCE($1, kode),
		    nama_perusahaan = COALESCE($2, nama_perusahaan),
		    tanggal_pencatatan = COALESCE($3, tanggal_pencatatan),
		    jumlah_saham = COALESCE($4, jumlah_saham),
		    papan_pencatatan = COALESCE($5, papan_pencatatan),
		    updated_at = NOW()
		WHERE id = $6
		RETURNING id, kode, nama_perusahaan, tanggal_pencatatan, jumlah_saham, papan_pencatatan, created_at, updated_at
	`

	var c model.Company
	err := r.db.QueryRow(ctx, query,
		company.Kode,
		company.NamaPerusahaan,
		company.TanggalPencatatan,
		company.JumlahSaham,
		company.PapanPencatatan,
		id,
	).Scan(
		&c.ID, &c.Kode, &c.NamaPerusahaan, &c.TanggalPencatatan,
		&c.JumlahSaham, &c.PapanPencatatan, &c.CreatedAt, &c.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *companyRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM companies WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("company not found")
	}

	return nil
}

func (r *companyRepository) BulkCreate(ctx context.Context, companies []model.Company) ([]model.Company, []error, error) {
	var created []model.Company
	var errors []error

	for _, company := range companies {
		query := `
			INSERT INTO companies (kode, nama_perusahaan, tanggal_pencatatan, jumlah_saham, papan_pencatatan, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			RETURNING id, kode, nama_perusahaan, tanggal_pencatatan, jumlah_saham, papan_pencatatan, created_at, updated_at
		`

		var c model.Company
		err := r.db.QueryRow(ctx, query,
			company.Kode,
			company.NamaPerusahaan,
			company.TanggalPencatatan,
			company.JumlahSaham,
			company.PapanPencatatan,
		).Scan(
			&c.ID, &c.Kode, &c.NamaPerusahaan, &c.TanggalPencatatan,
			&c.JumlahSaham, &c.PapanPencatatan, &c.CreatedAt, &c.UpdatedAt,
		)

		if err != nil {
			errors = append(errors, fmt.Errorf("row with kode %s: %w", company.Kode, err))
		} else {
			created = append(created, c)
		}
	}

	return created, errors, nil
}
