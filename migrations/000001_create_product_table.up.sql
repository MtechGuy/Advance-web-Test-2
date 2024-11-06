CREATE TABLE products (
    product_id bigserial PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL,
    category text NOT NULL,
    image_url text NOT NULL,
    price text NOT NULL,
    average_rating DECIMAL(3, 2) DEFAULT 0.00, -- Average rating from reviews
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE reviews (
    review_id bigserial PRIMARY KEY,
    product_id INT REFERENCES products(product_id) ON DELETE CASCADE,
    author VARCHAR(255),
    rating FLOAT CHECK (rating BETWEEN 1 AND 5),
    review_text text NOT NULL,
    helpful_count INT DEFAULT 0,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);

CREATE OR REPLACE FUNCTION automatic_average_rating()
RETURNS TRIGGER AS $$
BEGIN
    -- Update the average rating of the product associated with the new review
    UPDATE products
    SET average_rating = (
        SELECT ROUND(CAST(AVG(rating) AS NUMERIC), 2)
        FROM reviews
        WHERE reviews.product_id = NEW.product_id
    )
    WHERE product_id = NEW.product_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger that executes automatic_average_rating() when a review is added, updated, or deleted
CREATE OR REPLACE TRIGGER update_product_rating
AFTER INSERT OR UPDATE OR DELETE ON reviews
FOR EACH ROW
EXECUTE FUNCTION automatic_average_rating();
