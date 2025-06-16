-- Create gf_migrations table to track migrations, even if applied manually for now
CREATE TABLE IF NOT EXISTS gf_migrations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    batch INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Create test_migrations table (from our first migration attempt)
CREATE TABLE IF NOT EXISTS test_migrations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Record the 'CreateInitialTables' migration as if it were run by GoFrame
-- Use a naming convention similar to what GoFrame would generate for the file if possible
-- e.g., m_YYYYMMDDHHMMSS_CreateInitialTables
-- For simplicity, we'll use the name from the manually created .go file
INSERT IGNORE INTO gf_migrations (name, batch) VALUES ('m_20240101000000_CreateInitialTables', 1);
