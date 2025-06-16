-- Create posts table (user positions/job titles)
CREATE TABLE IF NOT EXISTS posts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE COMMENT 'Position code',
    name VARCHAR(100) NOT NULL COMMENT 'Position name',
    status TINYINT DEFAULT 1 COMMENT '1:active, 2:inactive',
    sort_order INT DEFAULT 0 COMMENT 'Sort order',
    remark VARCHAR(255) NULL COMMENT 'Optional remarks',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Stores user job positions/posts';

-- Record this migration
INSERT IGNORE INTO gf_migrations (name, batch) VALUES ('0003_create_posts_table', 4);
-- Assuming batch 3 was departments
