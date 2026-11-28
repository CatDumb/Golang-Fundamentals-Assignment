// Package repositories contains the data-access layer (SQL queries).
package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"product-api/internal/models"
)

// ErrNotFound is returned when a row with the requested id does not exist.
var ErrNotFound = errors.New("record not found")

// ProductRepository defines the data operations on products.
type ProductRepository interface {
	Create(ctx context.Context, p *models.Product) error
	FindAll(ctx context.Context, keyword string) ([]models.Product, error)
	FindByID(ctx context.Context, id int64) (*models.Product, error)
	Update(ctx context.Context, p *models.Product) error
	Delete(ctx context.Context, id int64) error
}

type postgresProductRepository struct {
	db *pgxpool.Pool
}

// NewProductRepository returns a PostgreSQL-backed ProductRepository.
func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &postgresProductRepository{db: db}
}

const productColumns = `id, name, COALESCE(description, ''), price::float8, quantity, created_at, updated_at`

func scanProduct(row pgx.Row, p *models.Product) error {
	return row.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Quantity, &p.CreatedAt, &p.UpdatedAt)
}

func (r *postgresProductRepository) Create(ctx context.Context, p *models.Product) error {
	query := `INSERT INTO products (name, description, price, quantity)
	          VALUES ($1, $2, $3, $4)
	          RETURNING ` + productColumns
	return scanProduct(r.db.QueryRow(ctx, query, p.Name, p.Description, p.Price, p.Quantity), p)
}

func (r *postgresProductRepository) FindAll(ctx context.Context, keyword string) ([]models.Product, error) {
	query := `SELECT ` + productColumns + ` FROM products`
	var args []any
	if keyword != "" {
		query += ` WHERE name ILIKE $1 OR description ILIKE $1`
		args = append(args, "%"+keyword+"%")
	}
	query += ` ORDER BY id`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var p models.Product
		if err := scanProduct(rows, &p); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *postgresProductRepository) FindByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `SELECT ` + productColumns + ` FROM products WHERE id = $1`
	var p models.Product
	if err := scanProduct(r.db.QueryRow(ctx, query, id), &p); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *postgresProductRepository) Update(ctx context.Context, p *models.Product) error {
	query := `UPDATE products
	          SET name = $1, description = $2, price = $3, quantity = $4, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $5
	          RETURNING ` + productColumns
	err := scanProduct(r.db.QueryRow(ctx, query, p.Name, p.Description, p.Price, p.Quantity, p.ID), p)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *postgresProductRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
