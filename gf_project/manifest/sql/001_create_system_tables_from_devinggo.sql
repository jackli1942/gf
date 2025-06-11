-- Extracted DDL from ./resource/migrations/20240628082802_initTable.up.sql

CREATE TABLE `system_user` (
                               `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                               `username` TEXT NOT NULL,
                               `password` TEXT NOT NULL,
                               `user_type` TEXT DEFAULT '100',
                               `nickname` TEXT DEFAULT NULL,
                               `phone` TEXT DEFAULT NULL,
                               `email` TEXT DEFAULT NULL,
                               `avatar` TEXT DEFAULT NULL,
                               `signed` TEXT DEFAULT NULL,
                               `dashboard` TEXT DEFAULT NULL,
                               `status` INTEGER DEFAULT 1,
                               `login_ip` TEXT DEFAULT NULL,
                               `login_time` DATETIME DEFAULT NULL,
                               `backend_setting` TEXT DEFAULT NULL,
                               `created_by` INTEGER DEFAULT NULL,
                               `updated_by` INTEGER DEFAULT NULL,
                               `created_at` DATETIME DEFAULT NULL,
                               `updated_at` DATETIME DEFAULT NULL,
                               `deleted_at` DATETIME DEFAULT NULL,
                               `remark` TEXT DEFAULT NULL,
                               UNIQUE (`username`)
);

CREATE TABLE `system_role` (
                               `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                               `name` TEXT NOT NULL,
                               `code` TEXT NOT NULL,
                               `data_scope` INTEGER DEFAULT 1,
                               `status` INTEGER DEFAULT 1,
                               `sort` INTEGER DEFAULT 0,
                               `created_by` INTEGER DEFAULT NULL,
                               `updated_by` INTEGER DEFAULT NULL,
                               `created_at` DATETIME DEFAULT NULL,
                               `updated_at` DATETIME DEFAULT NULL,
                               `deleted_at` DATETIME DEFAULT NULL,
                               `remark` TEXT DEFAULT NULL
);

CREATE TABLE `system_menu` (
                               `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                               `parent_id` INTEGER NOT NULL,
                               `level` TEXT NOT NULL,
                               `name` TEXT NOT NULL,
                               `code` TEXT NOT NULL,
                               `icon` TEXT DEFAULT NULL,
                               `route` TEXT DEFAULT NULL,
                               `component` TEXT DEFAULT NULL,
                               `redirect` TEXT DEFAULT NULL,
                               `is_hidden` INTEGER NOT NULL DEFAULT 1,
                               `type` TEXT NOT NULL DEFAULT '',
                               `status` INTEGER DEFAULT 1,
                               `sort` INTEGER DEFAULT 0,
                               `created_by` INTEGER DEFAULT NULL,
                               `updated_by` INTEGER DEFAULT NULL,
                               `created_at` DATETIME DEFAULT NULL,
                               `updated_at` DATETIME DEFAULT NULL,
                               `deleted_at` DATETIME DEFAULT NULL,
                               `remark` TEXT DEFAULT NULL
);

CREATE TABLE `system_dept` (
                               `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                               `parent_id` INTEGER NOT NULL,
                               `level` TEXT NOT NULL,
                               `name` TEXT NOT NULL,
                               `leader` TEXT DEFAULT NULL,
                               `phone` TEXT DEFAULT NULL,
                               `status` INTEGER DEFAULT 1,
                               `sort` INTEGER DEFAULT 0,
                               `created_by` INTEGER DEFAULT NULL,
                               `updated_by` INTEGER DEFAULT NULL,
                               `created_at` DATETIME DEFAULT NULL,
                               `updated_at` DATETIME DEFAULT NULL,
                               `deleted_at` DATETIME DEFAULT NULL,
                               `remark` TEXT DEFAULT NULL
);
-- SQLite does not support inline KEY definitions other than PRIMARY KEY and UNIQUE.
-- CREATE INDEX system_dept_parent_id_index ON system_dept (parent_id); -- If needed

CREATE TABLE `system_post` (
                               `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                               `name` TEXT NOT NULL,
                               `code` TEXT NOT NULL,
                               `sort` INTEGER DEFAULT 0,
                               `status` INTEGER DEFAULT 1,
                               `created_by` INTEGER DEFAULT NULL,
                               `updated_by` INTEGER DEFAULT NULL,
                               `created_at` DATETIME DEFAULT NULL,
                               `updated_at` DATETIME DEFAULT NULL,
                               `deleted_at` DATETIME DEFAULT NULL,
                               `remark` TEXT DEFAULT NULL
);

CREATE TABLE `system_user_role` (
                                    `user_id` INTEGER NOT NULL,
                                    `role_id` INTEGER NOT NULL,
                                    PRIMARY KEY (`user_id`,`role_id`)
);

CREATE TABLE `system_role_menu` (
                                    `role_id` INTEGER NOT NULL,
                                    `menu_id` INTEGER NOT NULL,
                                    PRIMARY KEY (`role_id`,`menu_id`)
);

CREATE TABLE `system_user_dept` (
                                    `user_id` INTEGER NOT NULL,
                                    `dept_id` INTEGER NOT NULL,
                                    PRIMARY KEY (`user_id`,`dept_id`)
);

CREATE TABLE `system_user_post` (
                                    `user_id` INTEGER NOT NULL,
                                    `post_id` INTEGER NOT NULL,
                                    PRIMARY KEY (`user_id`,`post_id`)
);

-- NOTE: Using 'casbin_rule' table for Casbin policies (singular name for gf-casbin/adapter compatibility).
-- Adding standard casbin_rule table for SQLite.

-- Casbin Rule Table: casbin_rule
CREATE TABLE IF NOT EXISTS `casbin_rule` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `ptype` TEXT DEFAULT NULL,
  `v0` TEXT DEFAULT NULL,
  `v1` TEXT DEFAULT NULL,
  `v2` TEXT DEFAULT NULL,
  `v3` TEXT DEFAULT NULL,
  `v4` TEXT DEFAULT NULL,
  `v5` TEXT DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS `idx_casbin_rule` ON `casbin_rule` (`ptype`, `v0`, `v1`);
