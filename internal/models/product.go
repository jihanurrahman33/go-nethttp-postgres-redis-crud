package models

import (
	"errors"
	"strings"
	"time"
)

// Product represents a product entity stored in Postgres and cached in Redis.
type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest defines the required payload to create a new product.
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

// Validate checks request validity for product creation.
func (r *CreateProductRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if r.Price < 0 {
		return errors.New("price cannot be negative")
	}
	if r.Stock < 0 {
		return errors.New("stock cannot be negative")
	}
	return nil
}

// UpdateProductRequest defines the payload to update an existing product.
type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

// Validate checks request validity for product update.
func (r *UpdateProductRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if r.Price < 0 {
		return errors.New("price cannot be negative")
	}
	if r.Stock < 0 {
		return errors.New("stock cannot be negative")
	}
	return nil
}
