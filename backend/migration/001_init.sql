-- DouTok 数据库初始化
-- 亿级用户架构设计：所有 ID 使用 Snowflake 算法生成 (int64)
-- 分库分表预留：表名含分片键注释，后续可用 ShardingSphere 拆分

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- ==================== 用户域 ====================

CREATE TABLE IF NOT EXISTS `users` (
    `id`            BIGINT       NOT NULL COMMENT 'Snowflake ID',
    `username`      VARCHAR(32)  NOT NULL,
    `phone`         VARCHAR(20)  DEFAULT NULL,
    `email`         VARCHAR(128) DEFAULT NULL,
    `password`      VARCHAR(128) NOT NULL COMMENT 'bcrypt hashed',
    `nickname`      VARCHAR(64)  NOT NULL,
    `avatar`        VARCHAR(512) DEFAULT '',
    `bio`           VARCHAR(256) DEFAULT '',
    `gender`        TINYINT      DEFAULT 0 COMMENT '0=unknown,1=male,2=female',
    `birthday`      VARCHAR(10)  DEFAULT '',
    `status`        TINYINT      DEFAULT 1 COMMENT '1=active,2=banned,3=deleted',
    `role`          VARCHAR(16)  DEFAULT 'user',
    `follow_count`  BIGINT       DEFAULT 0,
    `fan_count`     BIGINT       DEFAULT 0,
    `like_count`    BIGINT       DEFAULT 0,
    `video_count`   BIGINT       DEFAULT 0,
    `created_at`    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_phone` (`phone`),
    KEY `idx_email` (`email`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='用户主表 | 分片键: id | 预估分 256 张表';

-- 关注关系表：双写模型（关注方+被关注方各存一份，便于分片后本地查询）
CREATE TABLE IF NOT EXISTS `user_follows` (
    `id`         BIGINT   NOT NULL AUTO_INCREMENT,
    `user_id`    BIGINT   NOT NULL COMMENT '关注者',
    `target_id`  BIGINT   NOT NULL COMMENT '被关注者',
    `status`     TINYINT  DEFAULT 1 COMMENT '1=following,2=mutual,0=cancelled',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_target` (`user_id`, `target_id`),
    KEY `idx_target_id` (`target_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='关注关系表 | 分片键: user_id';

-- ==================== 视频域 ====================

CREATE TABLE IF NOT EXISTS `videos` (
    `id`             BIGINT       NOT NULL COMMENT 'Snowflake ID',
    `author_id`      BIGINT       NOT NULL,
    `title`          VARCHAR(128) NOT NULL,
    `description`    VARCHAR(512) DEFAULT '',
    `cover_url`      VARCHAR(512) NOT NULL,
    `play_url`       VARCHAR(512) NOT NULL,
    `duration`       INT          NOT NULL DEFAULT 0 COMMENT '秒',
    `width`          INT          DEFAULT 0,
    `height`         INT          DEFAULT 0,
    `file_size`      BIGINT       DEFAULT 0 COMMENT '字节',
    `status`         TINYINT      DEFAULT 0 COMMENT '0=processing,1=published,2=banned,3=deleted',
    `visibility`     TINYINT      DEFAULT 1 COMMENT '1=public,2=friends,3=private',
    `like_count`     BIGINT       DEFAULT 0,
    `comment_count`  BIGINT       DEFAULT 0,
    `share_count`    BIGINT       DEFAULT 0,
    `view_count`     BIGINT       DEFAULT 0,
    `tags`           VARCHAR(512) DEFAULT '' COMMENT '逗号分隔标签',
    `created_at`     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    KEY `idx_author_id` (`author_id`),
    KEY `idx_status_created` (`status`, `created_at` DESC),
    KEY `idx_created_at` (`created_at` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='视频主表 | 分片键: id | 冷热分离: 30天以上归档';

-- 点赞表：用 Redis ZSET 做实时计数，MySQL 做持久化和对账
CREATE TABLE IF NOT EXISTS `video_likes` (
    `id`         BIGINT  NOT NULL AUTO_INCREMENT,
    `video_id`   BIGINT  NOT NULL,
    `user_id`    BIGINT  NOT NULL,
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_video_user` (`video_id`, `user_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='视频点赞表 | 分片键: video_id | Redis ZSET 做实时计数';

-- ==================== 评论域 ====================

CREATE TABLE IF NOT EXISTS `comments` (
    `id`          BIGINT       NOT NULL COMMENT 'Snowflake ID',
    `video_id`    BIGINT       NOT NULL,
    `user_id`     BIGINT       NOT NULL,
    `parent_id`   BIGINT       DEFAULT 0 COMMENT '父评论ID, 0=一级评论',
    `root_id`     BIGINT       DEFAULT 0 COMMENT '根评论ID, 便于楼中楼查询',
    `content`     VARCHAR(512) NOT NULL,
    `like_count`  BIGINT       DEFAULT 0,
    `reply_count` INT          DEFAULT 0,
    `status`      TINYINT      DEFAULT 1 COMMENT '1=visible,2=hidden,3=deleted',
    `ip_location` VARCHAR(32)  DEFAULT '' COMMENT 'IP属地',
    `created_at`  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    KEY `idx_video_id` (`video_id`, `created_at` DESC),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_root_id` (`root_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='评论表 | 分片键: video_id | 二级评论用 root_id 聚合';

CREATE TABLE IF NOT EXISTS `comment_likes` (
    `id`         BIGINT NOT NULL AUTO_INCREMENT,
    `comment_id` BIGINT NOT NULL,
    `user_id`    BIGINT NOT NULL,
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_comment_user` (`comment_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='评论点赞表';

-- ==================== 聊天域 ====================

CREATE TABLE IF NOT EXISTS `conversations` (
    `id`            BIGINT      NOT NULL COMMENT 'Snowflake ID',
    `type`          TINYINT     DEFAULT 1 COMMENT '1=private,2=group',
    `last_msg_id`   BIGINT      DEFAULT 0,
    `last_msg_at`   DATETIME(3) DEFAULT NULL,
    `member_count`  INT         DEFAULT 2,
    `created_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='会话表 | 分片键: id';

CREATE TABLE IF NOT EXISTS `conversation_members` (
    `id`              BIGINT  NOT NULL AUTO_INCREMENT,
    `conversation_id` BIGINT  NOT NULL,
    `user_id`         BIGINT  NOT NULL,
    `unread_count`    INT     DEFAULT 0,
    `last_read_msg`   BIGINT  DEFAULT 0 COMMENT '已读消息ID水位线',
    `muted`           TINYINT DEFAULT 0,
    `joined_at`       DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_conv_user` (`conversation_id`, `user_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='会话成员表 | 双查: 按 user_id 查我的会话, 按 conversation_id 查成员';

CREATE TABLE IF NOT EXISTS `messages` (
    `id`              BIGINT       NOT NULL COMMENT 'Snowflake ID, 天然有序',
    `conversation_id` BIGINT       NOT NULL,
    `sender_id`       BIGINT       NOT NULL,
    `msg_type`        TINYINT      DEFAULT 1 COMMENT '1=text,2=image,3=video,4=emoji,5=system',
    `content`         TEXT         NOT NULL,
    `extra`           JSON         DEFAULT NULL COMMENT '扩展字段: 图片URL/视频信息等',
    `status`          TINYINT      DEFAULT 1 COMMENT '1=sent,2=delivered,3=read,4=recalled',
    `created_at`      DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    KEY `idx_conv_created` (`conversation_id`, `created_at` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='消息表 | 分片键: conversation_id | Snowflake ID 保证时序';

-- ==================== 直播域 ====================

CREATE TABLE IF NOT EXISTS `live_rooms` (
    `id`           BIGINT       NOT NULL COMMENT 'Snowflake ID',
    `anchor_id`    BIGINT       NOT NULL COMMENT '主播用户ID',
    `title`        VARCHAR(128) NOT NULL,
    `cover_url`    VARCHAR(512) DEFAULT '',
    `stream_key`   VARCHAR(128) NOT NULL COMMENT 'SRS推流密钥',
    `status`       TINYINT      DEFAULT 0 COMMENT '0=created,1=live,2=ended',
    `viewer_count` INT          DEFAULT 0 COMMENT '当前观众数(Redis维护)',
    `peak_viewer`  INT          DEFAULT 0,
    `like_count`   BIGINT       DEFAULT 0,
    `gift_value`   BIGINT       DEFAULT 0 COMMENT '礼物总价值',
    `started_at`   DATETIME(3)  DEFAULT NULL,
    `ended_at`     DATETIME(3)  DEFAULT NULL,
    `created_at`   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    KEY `idx_anchor_id` (`anchor_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='直播间表';

CREATE TABLE IF NOT EXISTS `live_gifts` (
    `id`         BIGINT  NOT NULL AUTO_INCREMENT,
    `room_id`    BIGINT  NOT NULL,
    `sender_id`  BIGINT  NOT NULL,
    `gift_id`    INT     NOT NULL,
    `gift_name`  VARCHAR(32) NOT NULL,
    `gift_value` INT     NOT NULL COMMENT '礼物价值(虚拟币)',
    `count`      INT     DEFAULT 1,
    `combo`      INT     DEFAULT 1 COMMENT '连击数',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    KEY `idx_room_created` (`room_id`, `created_at` DESC),
    KEY `idx_sender_id` (`sender_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='直播礼物记录';

-- ==================== 搜索 & 推荐辅助表 ====================

-- 用户行为日志 - 推荐算法的数据源
-- 生产环境应使用 Kafka -> ClickHouse/Doris 链路
CREATE TABLE IF NOT EXISTS `user_behaviors` (
    `id`           BIGINT      NOT NULL AUTO_INCREMENT,
    `user_id`      BIGINT      NOT NULL,
    `target_type`  TINYINT     NOT NULL COMMENT '1=video,2=live,3=user',
    `target_id`    BIGINT      NOT NULL,
    `action`       VARCHAR(16) NOT NULL COMMENT 'view/like/comment/share/follow/gift',
    `duration`     INT         DEFAULT 0 COMMENT '观看时长(秒)',
    `extra`        JSON        DEFAULT NULL,
    `created_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    KEY `idx_user_created` (`user_id`, `created_at` DESC),
    KEY `idx_target` (`target_type`, `target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='用户行为日志 | 生产环境: Kafka->ClickHouse';
