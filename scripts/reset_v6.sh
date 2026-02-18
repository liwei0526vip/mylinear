#!/bin/bash
# 自动修复数据库 v6 迁移失败残留
# 设置颜色
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}正在清理数据库 v6 迁移残留...${NC}"

# 检查 Docker 容器是否运行
if ! docker compose ps | grep -q "postgres"; then
    echo -e "${RED}错误: PostgreSQL 容器未运行，请先执行 'make infra-up'${NC}"
    exit 1
fi

# 在容器内执行 SQL
docker exec -i mylinear-postgres psql -U mylinear -d mylinear <<SQL
-- 1. 清理 issues 表新增的列 (不管存不存在，尝试删掉)
ALTER TABLE issues DROP COLUMN IF EXISTS position;
ALTER TABLE issues DROP COLUMN IF EXISTS deleted_at;

-- 2. 删除新增的表
DROP TABLE IF EXISTS issue_subscriptions;

-- 3. 重置迁移版本号为 5 (强制设置为 dirty=false)
UPDATE schema_migrations SET version = 5, dirty = false;
SQL

if [ $? -eq 0 ]; then
    echo -e "${GREEN}清理完成！版本已重置为 v5。${NC}"
    echo -e "${GREEN}现在可以重新启动后端服务应用迁移了。${NC}"
else
    echo -e "${RED}SQL 执行失败。${NC}"
    exit 1
fi
