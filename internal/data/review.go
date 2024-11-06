// Filename: internal/data/reviews.go
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mtechguy/test2/internal/validator"
)

// Review struct
type Review struct {
	ReviewID     int64     `json:"review_id"`  // bigserial primary key
	ProductID    int64     `json:"product_id"` // foreign key referencing products
	Author       string    `json:"author"`
	Rating       int64     `json:"rating"`        // integer with a constraint (1-5)
	ReviewText   string    `json:"review_text"`   // non-null text field
	HelpfulCount int32     `json:"helpful_count"` // nullable integer, default 0
	CreatedAt    time.Time `json:"-"`             // timestamp with timezone, default now()
	Version      int       `json:"version"`
}

type ReviewModel struct {
	DB *sql.DB
}

func ValidateReview(v *validator.Validator, review *Review) {
	v.Check(review.Author != "", "author", "must be provided")
	v.Check(review.ReviewText != "", "review_text", "must be provided")

	v.Check(len(review.Author) <= 25, "author", "must not be more than 25 bytes long")
	v.Check(review.ProductID > 0, "product_id", "must be a positive integer")
	v.Check(review.Rating >= 1 && review.Rating <= 5, "rating", "must be between 1 and 5")
}

func (c ReviewModel) InsertReview(review *Review) error {
	query := `
		INSERT INTO reviews (product_id, author, rating, review_text, helpful_count)
		VALUES ($1, $2, $3, $4, COALESCE($5, 0))
		RETURNING review_id, created_at, version
	`
	args := []any{review.ProductID, review.Author, review.Rating, review.ReviewText, review.HelpfulCount}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&review.ReviewID,
		&review.CreatedAt,
		&review.Version)
}
func (c ReviewModel) GetReview(id int64) (*Review, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT review_id, product_id, author, rating, review_text, helpful_count, created_at, version
		FROM reviews
		WHERE review_id = $1
	`
	var review Review

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&review.ReviewID,
		&review.ProductID,
		&review.Author,
		&review.Rating,
		&review.ReviewText,
		&review.HelpfulCount,
		&review.CreatedAt,
		&review.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &review, nil
}

func (c ReviewModel) UpdateReview(review *Review) error {
	query := `
		UPDATE reviews
		SET author = $1, rating = $2, review_text = $3, version = version + 1
		WHERE review_id = $4
		RETURNING version
	`

	args := []any{review.Author, review.Rating, review.ReviewText, review.ReviewID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&review.Version)
}

func (c ReviewModel) DeleteReview(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
		DELETE FROM reviews
		WHERE review_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.DB.ExecContext(ctx, query, id)
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

func (c ReviewModel) GetAllReviews(author string, filters Filters) ([]*Review, Metadata, error) {
	// Construct the SQL query with placeholders for parameters
	query := fmt.Sprintf(`
	SELECT COUNT(*) OVER(), review_id, product_id, author, rating, review_text, helpful_count, created_at, version
	FROM reviews
	WHERE (to_tsvector('simple', author) @@ plainto_tsquery('simple', $1) OR $1 = '') 
	ORDER BY %s %s, review_id ASC 
	LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	// Set a context with a 3-second timeout for query execution
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query with provided filters and parameters
	rows, err := c.DB.QueryContext(ctx, query, author, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var totalRecords int
	reviews := []*Review{}

	// Iterate over result rows and scan data into Review struct
	for rows.Next() {
		var review Review
		if err := rows.Scan(&totalRecords, &review.ReviewID, &review.ProductID, &review.Author, &review.Rating, &review.ReviewText, &review.HelpfulCount, &review.CreatedAt, &review.Version); err != nil {
			return nil, Metadata{}, err
		}
		reviews = append(reviews, &review)
	}

	// Check if any error occurred during row iteration
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Calculate metadata for pagination
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return reviews, metadata, nil
}

func (c ReviewModel) GetAllProductReviews(productID int64) ([]Review, error) {
	if productID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT review_id, author, rating, review_text, helpful_count, created_at, version
		FROM reviews
		WHERE product_id = $1
	`

	// Initialize a slice to hold all reviews for the product
	var reviews []Review

	// Set up the context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Query all rows that match the productID
	rows, err := c.DB.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the rows and scan each row into a Review struct
	for rows.Next() {
		var review Review
		err := rows.Scan(
			&review.ReviewID,
			&review.Author,
			&review.Rating,
			&review.ReviewText,
			&review.HelpfulCount,
			&review.CreatedAt,
			&review.Version,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	// Check for any errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (c *ReviewModel) UpdateHelpfulCount(id int64) (*Review, error) {
	query := `
        UPDATE reviews
        SET helpful_count = helpful_count + 1
        WHERE review_id = $1
        RETURNING review_id, author, rating, review_text, helpful_count, version
    `

	var review Review
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query and scan the updated review fields
	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&review.ReviewID,
		&review.Author,
		&review.Rating,
		&review.ReviewText,
		&review.HelpfulCount,
		&review.Version,
	)
	if err != nil {
		return nil, err
	}

	return &review, nil
}

func (m *ProductModel) ProductExists(productID int64) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM products WHERE product_id = $1)`
	var exists bool
	err := m.DB.QueryRow(query, productID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
func (m *ReviewModel) Exists(id int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM reviews WHERE review_id = $1)`
	err := m.DB.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c ReviewModel) GetProductReview(rid int64, pid int64) (*Review, error) {
	//validate id
	if pid < 1 || rid < 1 {
		return nil, ErrRecordNotFound
	}

	//query
	query := `SELECT review_id, product_id, author, rating, review_text, helpful_count, created_at, version
	FROM reviews
	WHERE review_id = $1 AND product_id = $2
	`
	var review Review

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, rid, pid).Scan(
		&review.ReviewID,
		&review.ProductID,
		&review.Author,
		&review.Rating,
		&review.ReviewText,
		&review.HelpfulCount,
		&review.CreatedAt,
		&review.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &review, nil
}
