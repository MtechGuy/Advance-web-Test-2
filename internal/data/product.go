// Filename: internal/data/products.go
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mtechguy/test2/internal/validator"
)

type Product struct {
	ProductID     int64     `json:"product_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	ImageURL      string    `json:"image_url"`
	Price         string    `json:"price"`
	AverageRating float32   `json:"average_rating"`
	CreatedAt     time.Time `json:"-"`
	Version       int32     `json:"version"`
}

type ProductModel struct {
	DB *sql.DB
}

// Validation function for Product struct
func ValidateProduct(v *validator.Validator, product *Product) {
	v.Check(product.Name != "", "name", "must be provided")
	v.Check(len(product.Name) <= 100, "name", "must not be more than 100 characters long")
	v.Check(product.Description != "", "description", "must be provided")
	v.Check(len(product.Description) <= 500, "description", "must not be more than 500 characters long")
	v.Check(product.Category != "", "category", "must be provided")
	v.Check(product.ImageURL != "", "image_url", "must be provided")
	v.Check(len(product.ImageURL) <= 255, "image_url", "must not be more than 255 characters long")
	v.Check(len(product.Price) <= 10, "price", "must not be more than 10 characters long")
	v.Check(product.Description != "", "description", "must be provided")
	// v.Check(product.AverageRating >= 0 && product.AverageRating <= 5, "average_rating", "must be between 0 and 5")
}

func (p ProductModel) InsertProduct(product *Product) error {
	query := `
		INSERT INTO products (name, description, category, image_url, price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING product_id, created_at, version
	`
	args := []any{product.Name, product.Description, product.Category, product.ImageURL, product.Price}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, args...).Scan(
		&product.ProductID,
		&product.CreatedAt,
		&product.Version,
	)
}

func (p ProductModel) GetProduct(id int64) (*Product, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT product_id, name, description, category, image_url, price, average_rating, created_at, version
		FROM products
		WHERE product_id = $1
	`

	var product Product
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, id).Scan(
		&product.ProductID,
		&product.Name,
		&product.Description,
		&product.Category,
		&product.ImageURL,
		&product.Price,
		&product.AverageRating,
		&product.CreatedAt,
		&product.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &product, nil
}

func (p ProductModel) UpdateProduct(product *Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, category = $3, image_url = $4, price = $5, average_rating = $6, version = version + 1
		WHERE product_id = $7
		RETURNING version
	`

	// Removed `product.UpdatedAt` from the args slice
	args := []any{product.Name, product.Description, product.Category, product.ImageURL, product.Price, product.AverageRating, product.ProductID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, args...).Scan(&product.Version)
}

func (p ProductModel) DeleteProduct(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM products
		WHERE product_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (p ProductModel) GetAllProducts(name string, category string, filters Filters) ([]*Product, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), product_id, name, description, category, image_url, price, average_rating, created_at, version
		FROM products
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '') 
		AND (to_tsvector('simple', category) @@ plainto_tsquery('simple', $2) OR $2 = '') 
		ORDER BY %s %s, product_id ASC 
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, name, category, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0
	products := []*Product{}

	for rows.Next() {
		var product Product
		err := rows.Scan(
			&totalRecords,
			&product.ProductID,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.ImageURL,
			&product.Price,
			&product.AverageRating,
			&product.CreatedAt,
			&product.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		products = append(products, &product)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return products, metadata, nil
}
