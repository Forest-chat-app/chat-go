-- 数据库名称：chat (这里是创建一个名为 chat 的数据库，如果你希望使用现有数据库，可以移除这行，并在 Go 配置中指定数据库名)
CREATE DATABASE IF NOT EXISTS `chat`;
USE `chat`;

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


CREATE TABLE IF NOT EXISTS `room` (
    `id` VARCHAR(255) NOT NULL COMMENT '房间ID',
    `room_name` VARCHAR(255) NOT NULL,
    `creator_id` VARCHAR(255) NOT NULL,
    `is_private` TINYINT(1) NOT NULL COMMENT '是否私有，0为否，1为是', -- JSON 类型在 MySQL 中通常用于存储复杂结构，对于布尔值建议使用 TINYINT(1)
    `is_delete` TINYINT(1) NOT NULL COMMENT '是否删除，0为否，1为是', -- 同上
    `created_at` BIGINT NOT NULL COMMENT '创建时间戳 (毫秒)', -- INTEGER 类型通常是秒，这里改为 BIGINT 假设是毫秒
    `updated_at` BIGINT NOT NULL COMMENT '更新时间戳 (毫秒)', -- 同上
    PRIMARY KEY (`id`),
    FOREIGN KEY (`creator_id`) REFERENCES `user`(`id`) ON DELETE CASCADE ON UPDATE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE IF NOT EXISTS `room_members` (
    `id` VARCHAR(255) NOT NULL COMMENT 'ID 编号',
    `user_id` VARCHAR(255) NOT NULL,
    `room_id` VARCHAR(255) NOT NULL,
    `joined_at` BIGINT NOT NULL COMMENT '加入时间戳 (毫秒)',
    PRIMARY KEY (`id`),
    FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (`room_id`) REFERENCES `room`(`id`) ON DELETE CASCADE ON UPDATE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
