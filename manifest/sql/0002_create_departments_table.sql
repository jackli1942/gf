-- Create departments table
CREATE TABLE IF NOT EXISTS departments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    parent_id INT DEFAULT 0 COMMENT 'Parent department ID, 0 for root',
    name VARCHAR(100) NOT NULL,
    leader VARCHAR(50) NULL,
    status TINYINT DEFAULT 1 COMMENT '1:active, 2:inactive',
    sort_order INT DEFAULT 0 COMMENT 'Sort order',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_parent_id (parent_id),
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Stores department information';

-- Record this migration
INSERT IGNORE INTO gf_migrations (name, batch) VALUES ('0002_create_departments_table', 3);
-- Assuming batch 1 was initial_tables, batch 2 was users table
