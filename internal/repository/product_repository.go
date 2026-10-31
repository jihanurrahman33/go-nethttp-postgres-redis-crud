package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-crud-backend/internal/models"
)

var (
	// ErrNotFound is returned when an entity is not found in the database.
	ErrNotFound = errors.New("record not found")
)

// ProductRepository outlines database interactions for products.
type ProductRepository interface {
	Create(ctx context.Context, p *models.Product) error
	GetByID(ctx context.Context, id int64) (*models.Product, error)
	GetAll(ctx context.Context, limit, offset int) ([]models.Product, error)
	Update(ctx context.Context, p *models.Product) error
	Delete(ctx context.Context, id int64) error
}

type postgresProductRepo struct {
	db *sql.DB
}

// NewPostgresProductRepository constructs a new PostgreSQL-backed product repository.
func NewPostgresProductRepository(db *sql.DB) ProductRepository {
	return &postgresProductRepo{db: db}
}

func (r *postgresProductRepo) Create(ctx context.Context, p *models.Product) error {
	query := `
		INSERT INTO products (name, description, price, stock, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	now := time.Now().UTC()
	err := r.db.QueryRowContext(
		ctx,
		query,
		p.Name,
		p.Description,
		p.Price,
		p.Stock,
		now,
		now,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert product: %w", err)
	}

	return nil
}

func (r *postgresProductRepo) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	var p models.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Price,
		&p.Stock,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query product: %w", err)
	}

	return &p, nil
}

func (r *postgresProductRepo) GetAll(ctx context.Context, limit, offset int) ([]models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, created_at, updated_at
		FROM products
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Stock,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading products rows: %w", err)
	}

	return products, nil
}

func (r *postgresProductRepo) Update(ctx context.Context, p *models.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4, updated_at = $5
		WHERE id = $6
		RETURNING updated_at
	`
	now := time.Now().UTC()
	err := r.db.QueryRowContext(
		ctx,
		query,
		p.Name,
		p.Description,
		p.Price,
		p.Stock,
		now,
		p.ID,
	).Scan(&p.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

func (r *postgresProductRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
