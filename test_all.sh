#!/bin/bash

# MigraGuard 전체 DDL 분석 테스트 스크립트 (Linux/Container)

DB_URL="postgres://user:pass@db:5432/target_db"
SQLITE_PATH="/app/data/migraguard.db"
MIGRATIONS_DIR="/app/code/migrations"

echo "🚀 모든 DDL 파일에 대한 리스크 분석을 시작합니다..."

for file in "$MIGRATIONS_DIR"/*.sql; do
    fileName=$(basename "$file")
    echo ""
    echo "[파일 분석 중: $fileName]"
    
    # 내부 바이너리(migraguard) 직접 호출
    migraguard analyze "$file" --db "$DB_URL" --sqlite "$SQLITE_PATH"
done

echo ""
echo "✅ 모든 분석이 완료되었습니다."
