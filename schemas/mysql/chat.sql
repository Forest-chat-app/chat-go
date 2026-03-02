-- 数据库名称：chat (这里是创建一个名为 chat 的数据库，如果你希望使用现有数据库，可以移除这行，并在 Go 配置中指定数据库名)
CREATE DATABASE IF NOT EXISTS `chat`;
USE `chat`;

-- DROP TABLE IF EXISTS `user`;
CREATE TABLE IF NOT EXISTS `user` (
    `id` VARCHAR(255) NOT NULL COMMENT '用户ID',
    `user_account` VARCHAR(255) NOT NULL UNIQUE, -- user_account 通常是唯一的
    `password` VARCHAR(255) NOT NULL,
    `nickname` VARCHAR(255) NOT NULL,
    `avatar` VARCHAR(255) NOT NULL,
    `email` VARCHAR(255) NOT NULL,
    `created_at` BIGINT NOT NULL COMMENT '创建时间戳 (毫秒)',
    `updated_at` BIGINT NOT NULL COMMENT '更新时间戳 (毫秒)',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- DROP TABLE IF EXISTS `role`;
CREATE TABLE IF NOT EXISTS `role` (
    `id` VARCHAR(255) NOT NULL COMMENT '角色id',
    `name` VARCHAR(255) NOT NULL COMMENT '角色名称',
    `description` VARCHAR(255) NOT NULL COMMENT '角色描述',
    `created_at` BIGINT NOT NULL COMMENT '创建时间',
    `updated_at` BIGINT NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- DROP TABLE IF EXISTS `user_role`;
CREATE TABLE IF NOT EXISTS `user_role` (
    `id` VARCHAR(255) NOT NULL COMMENT '用户角色id',
    `user_id` VARCHAR(255) NOT NULL COMMENT '用户id',
    `role_id` VARCHAR(255) NOT NULL COMMENT '角色id',
    `created_at` BIGINT NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (`role_id`) REFERENCES `role`(`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- DROP TABLE IF EXISTS `room`;
CREATE TABLE IF NOT EXISTS `room` (
    `id` VARCHAR(255) NOT NULL COMMENT 'ID 编号',
    `creator_id` VARCHAR(255) NOT NULL COMMENT '创建人id',
    `room_name` VARCHAR(255) NOT NULL COMMENT '聊天室名称',
    `introduction` VARCHAR(255) NOT NULL COMMENT '房间简介',
    `tag` VARCHAR(255) NOT NULL COMMENT '标签，以#号隔开',
    `status` TINYINT NOT NULL COMMENT '1、公开 2、私密 3、已删除',
    `created_at` BIGINT NOT NULL COMMENT '创建时间 毫秒',
    `updated_at` BIGINT NOT NULL COMMENT '更新时间 毫秒',
    PRIMARY KEY (`id`),
    FOREIGN KEY (`creator_id`) REFERENCES `user`(`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- DROP TABLE IF EXISTS `room_members`;
CREATE TABLE IF NOT EXISTS `room_members` (
    `id` VARCHAR(255) NOT NULL COMMENT 'ID 编号',
    `user_id` VARCHAR(255) NOT NULL,
    `room_id` VARCHAR(255) NOT NULL,
    `joined_at` BIGINT NOT NULL COMMENT '加入时间戳 (毫秒)',
    PRIMARY KEY (`id`),
    FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (`room_id`) REFERENCES `room`(`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
