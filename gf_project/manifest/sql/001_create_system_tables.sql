-- Users Table: sys_user
CREATE TABLE IF NOT EXISTS `sys_user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'User ID',
  `username` VARCHAR(60) NOT NULL COMMENT 'Username',
  `password_hash` VARCHAR(100) NOT NULL COMMENT 'Hashed Password',
  `nickname` VARCHAR(60) DEFAULT NULL COMMENT 'Nickname',
  `avatar_url` VARCHAR(255) DEFAULT '' COMMENT 'User Avatar URL',
  `gender` TINYINT UNSIGNED DEFAULT 0 COMMENT 'Gender (0: Unknown, 1: Male, 2: Female)',
  `email` VARCHAR(100) DEFAULT NULL COMMENT 'Email Address',
  `mobile` VARCHAR(20) DEFAULT NULL COMMENT 'Mobile Phone Number',
  `dept_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'Department ID',
  `status` TINYINT UNSIGNED DEFAULT 0 COMMENT 'Status (0: Active, 1: Disabled)',
  `remark` VARCHAR(255) DEFAULT NULL COMMENT 'Remarks',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created Timestamp',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Updated Timestamp',
  `deleted_at` DATETIME DEFAULT NULL COMMENT 'Deleted Timestamp (for soft deletes)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_user_username` (`username`),
  UNIQUE KEY `uk_sys_user_email` (`email`),
  UNIQUE KEY `uk_sys_user_mobile` (`mobile`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='System Users Table';

-- Roles Table: sys_role
CREATE TABLE IF NOT EXISTS `sys_role` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Role ID',
  `name` VARCHAR(60) NOT NULL COMMENT 'Role Name',
  `role_key` VARCHAR(60) NOT NULL COMMENT 'Role Key (for permissions, e.g., admin, user)',
  `remark` VARCHAR(255) DEFAULT NULL COMMENT 'Remarks',
  `status` TINYINT UNSIGNED DEFAULT 0 COMMENT 'Status (0: Active, 1: Disabled)',
  `sort_order` INT DEFAULT 0 COMMENT 'Sort Order',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created Timestamp',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Updated Timestamp',
  `deleted_at` DATETIME DEFAULT NULL COMMENT 'Deleted Timestamp',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_role_role_key` (`role_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='System Roles Table';

-- Menus Table: sys_menu
CREATE TABLE IF NOT EXISTS `sys_menu` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Menu ID',
  `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Parent Menu ID (0 for top-level)',
  `name` VARCHAR(60) NOT NULL COMMENT 'Menu Name/Key (for i18n or code)',
  `title` VARCHAR(100) NOT NULL COMMENT 'Menu Title (display name)',
  `path` VARCHAR(255) DEFAULT NULL COMMENT 'Routing Path (frontend)',
  `component` VARCHAR(255) DEFAULT NULL COMMENT 'Component Path (frontend)',
  `icon` VARCHAR(100) DEFAULT NULL COMMENT 'Menu Icon',
  `menu_type` CHAR(1) NOT NULL COMMENT 'Menu Type (M: Directory, C: Menu, F: Button/Function)',
  `permission` VARCHAR(100) DEFAULT NULL COMMENT 'Permission Identifier (e.g., system:user:list)',
  `status` TINYINT UNSIGNED DEFAULT 0 COMMENT 'Status (0: Active, 1: Disabled)',
  `sort_order` INT DEFAULT 0 COMMENT 'Sort Order',
  `visible` TINYINT UNSIGNED DEFAULT 1 COMMENT 'Visible in menu (0: No, 1: Yes)',
  `breadcrumb_visible` TINYINT UNSIGNED DEFAULT 1 COMMENT 'Breadcrumb Visible (0: No, 1: Yes)',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created Timestamp',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Updated Timestamp',
  `deleted_at` DATETIME DEFAULT NULL COMMENT 'Deleted Timestamp',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='System Menus Table';

-- Departments Table: sys_dept
CREATE TABLE IF NOT EXISTS `sys_dept` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Department ID',
  `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Parent Department ID (0 for top-level)',
  `name` VARCHAR(100) NOT NULL COMMENT 'Department Name',
  `ancestors` VARCHAR(255) DEFAULT '' COMMENT 'Ancestor IDs (e.g., ,1,2,)',
  `leader_user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'Department Leader User ID',
  `phone` VARCHAR(20) DEFAULT NULL COMMENT 'Phone Number',
  `email` VARCHAR(100) DEFAULT NULL COMMENT 'Email Address',
  `status` TINYINT UNSIGNED DEFAULT 0 COMMENT 'Status (0: Active, 1: Disabled)',
  `sort_order` INT DEFAULT 0 COMMENT 'Sort Order',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created Timestamp',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Updated Timestamp',
  `deleted_at` DATETIME DEFAULT NULL COMMENT 'Deleted Timestamp',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='System Departments Table';

-- Posts Table: sys_post
CREATE TABLE IF NOT EXISTS `sys_post` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Post ID',
  `code` VARCHAR(60) NOT NULL COMMENT 'Post Code',
  `name` VARCHAR(100) NOT NULL COMMENT 'Post Name',
  `remark` VARCHAR(255) DEFAULT NULL COMMENT 'Remarks',
  `status` TINYINT UNSIGNED DEFAULT 0 COMMENT 'Status (0: Active, 1: Disabled)',
  `sort_order` INT DEFAULT 0 COMMENT 'Sort Order',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created Timestamp',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Updated Timestamp',
  `deleted_at` DATETIME DEFAULT NULL COMMENT 'Deleted Timestamp',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_post_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='System Posts Table';

-- User-Role Join Table: sys_user_role
CREATE TABLE IF NOT EXISTS `sys_user_role` (
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'User ID',
  `role_id` BIGINT UNSIGNED NOT NULL COMMENT 'Role ID',
  PRIMARY KEY (`user_id`, `role_id`),
  CONSTRAINT `fk_sys_user_role_user_id` FOREIGN KEY (`user_id`) REFERENCES `sys_user` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_sys_user_role_role_id` FOREIGN KEY (`role_id`) REFERENCES `sys_role` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User-Role Relationships';

-- Role-Menu Join Table: sys_role_menu
CREATE TABLE IF NOT EXISTS `sys_role_menu` (
  `role_id` BIGINT UNSIGNED NOT NULL COMMENT 'Role ID',
  `menu_id` BIGINT UNSIGNED NOT NULL COMMENT 'Menu ID',
  PRIMARY KEY (`role_id`, `menu_id`),
  CONSTRAINT `fk_sys_role_menu_role_id` FOREIGN KEY (`role_id`) REFERENCES `sys_role` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_sys_role_menu_menu_id` FOREIGN KEY (`menu_id`) REFERENCES `sys_menu` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Role-Menu Relationships';

-- Casbin Rule Table: casbin_rule
CREATE TABLE IF NOT EXISTS `casbin_rule` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `ptype` VARCHAR(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `v0` VARCHAR(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v1` VARCHAR(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v2` VARCHAR(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v3` VARCHAR(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v4` VARCHAR(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v5` VARCHAR(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_casbin_rule` (`ptype`, `v0`, `v1`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Casbin Authorization Policies';
