-- Drop the trigger first as it depends on the function and table
DROP TRIGGER IF EXISTS update_product_rating ON reviews;

-- Drop the function next
DROP FUNCTION IF EXISTS automatic_average_rating();

-- Drop the tables, starting with the one that references another table
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS products;
