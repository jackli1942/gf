-- Extracted DDL from ./resource/migrations/20240628082802_initTable.up.sql

CREATE TABLE `system_user` (
                               `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '用户ID，主键',
                               `username` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户名',
                               `password` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码',
                               `user_type` varchar(3) COLLATE utf8mb4_unicode_ci DEFAULT '100' COMMENT '用户类型：(100系统用户)',
                               `nickname` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户昵称',
                               `phone` varchar(11) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机',
                               `email` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户邮箱',
                               `avatar` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户头像',
                               `signed` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '个人签名',
                               `dashboard` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '后台首页类型',
                               `status` smallint(6) DEFAULT '1' COMMENT '状态 (1正常 2停用)',
                               `login_ip` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '最后登陆IP',
                               `login_time` timestamp NULL DEFAULT NULL COMMENT '最后登陆时间',
                               `backend_setting` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '后台设置数据',
                               `created_by` bigint(20) DEFAULT NULL COMMENT '创建者',
                               `updated_by` bigint(20) DEFAULT NULL COMMENT '更新者',
                               `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
                               `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
                               `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
                               `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
                               PRIMARY KEY (`id`),
                               UNIQUE KEY `system_user_username_unique` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户信息表';

CREATE TABLE `system_role` (
                               `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
                               `name` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名称',
                               `code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色代码',
                               `data_scope` smallint(6) DEFAULT '1' COMMENT '数据范围（1：全部数据权限 2：自定义数据权限 3：本部门数据权限 4：本部门及以下数据权限 5：本人数据权限）',
                               `status` smallint(6) DEFAULT '1' COMMENT '状态 (1正常 2停用)',
                               `sort` smallint(5)  DEFAULT '0' COMMENT '排序',
                               `created_by` bigint(20) DEFAULT NULL COMMENT '创建者',
                               `updated_by` bigint(20) DEFAULT NULL COMMENT '更新者',
                               `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
                               `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
                               `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
                               `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
                               PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色信息表';

CREATE TABLE `system_menu` (
                               `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
                               `parent_id` bigint(20) NOT NULL COMMENT '父ID',
                               `level` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组级集合',
                               `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单名称',
                               `code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单标识代码',
                               `icon` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '菜单图标',
                               `route` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '路由地址',
                               `component` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '组件路径',
                               `redirect` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '跳转地址',
                               `is_hidden` smallint(6) NOT NULL DEFAULT '1' COMMENT '是否隐藏 (1是 2否)',
                               `type` char(1) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '菜单类型, (M菜单 B按钮 L链接 I iframe)',
                               `status` smallint(6) DEFAULT '1' COMMENT '状态 (1正常 2停用)',
                               `sort` smallint(5)  DEFAULT '0' COMMENT '排序',
                               `created_by` bigint(20) DEFAULT NULL COMMENT '创建者',
                               `updated_by` bigint(20) DEFAULT NULL COMMENT '更新者',
                               `created_at` timestamp NULL DEFAULT NULL,
                               `updated_at` timestamp NULL DEFAULT NULL,
                               `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
                               `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
                               PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4519 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单信息表';

CREATE TABLE `system_dept` (
                               `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
                               `parent_id` bigint(20) NOT NULL COMMENT '父ID',
                               `level` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组级集合',
                               `name` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '部门名称',
                               `leader` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '负责人',
                               `phone` varchar(11) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话',
                               `status` smallint(6) DEFAULT '1' COMMENT '状态 (1正常 2停用)',
                               `sort` smallint(5)  DEFAULT '0' COMMENT '排序',
                               `created_by` bigint(20) DEFAULT NULL COMMENT '创建者',
                               `updated_by` bigint(20) DEFAULT NULL COMMENT '更新者',
                               `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
                               `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
                               `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
                               `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
                               PRIMARY KEY (`id`),
                               KEY `system_dept_parent_id_index` (`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门信息表';

CREATE TABLE `system_post` (
                               `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
                               `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '岗位名称',
                               `code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '岗位代码',
                               `sort` smallint(5)  DEFAULT '0' COMMENT '排序',
                               `status` smallint(6) DEFAULT '1' COMMENT '状态 (1正常 2停用)',
                               `created_by` bigint(20) DEFAULT NULL COMMENT '创建者',
                               `updated_by` bigint(20) DEFAULT NULL COMMENT '更新者',
                               `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
                               `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
                               `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
                               `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
                               PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位信息表';

CREATE TABLE `system_user_role` (
                                    `user_id` bigint(20) NOT NULL COMMENT '用户主键',
                                    `role_id` bigint(20) NOT NULL COMMENT '角色主键',
                                    PRIMARY KEY (`user_id`,`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户与角色关联表';

CREATE TABLE `system_role_menu` (
                                    `role_id` bigint(20) NOT NULL COMMENT '角色主键',
                                    `menu_id` bigint(20) NOT NULL COMMENT '菜单主键',
                                    PRIMARY KEY (`role_id`,`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色与菜单关联表';

CREATE TABLE `system_user_dept` (
                                    `user_id` bigint(20) NOT NULL COMMENT '用户主键',
                                    `dept_id` bigint(20) NOT NULL COMMENT '部门主键',
                                    PRIMARY KEY (`user_id`,`dept_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户与部门关联表';

CREATE TABLE `system_user_post` (
                                    `user_id` bigint(20) NOT NULL COMMENT '用户主键',
                                    `post_id` bigint(20) NOT NULL COMMENT '岗位主键',
                                    PRIMARY KEY (`user_id`,`post_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户与岗位关联表';

-- NOTE: casbin_rule table not found in source schema.
