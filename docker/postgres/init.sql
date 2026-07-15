-- ITCFG 配置中台数据库初始化
-- 此文件在 PostgreSQL 容器首次启动时自动执行

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 创建数据库(如果不存在)
-- 注意：POSTGRES_DB 环境变量已创建 itcfg 数据库，此处仅作扩展配置

-- 输出信息
DO $$
BEGIN
    RAISE NOTICE 'ITCFG 数据库初始化完成';
END $$;