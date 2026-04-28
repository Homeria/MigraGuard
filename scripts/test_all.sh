#!/bin/bash

# MigraGuard 전체 DDL 분석 테스트 스크립트 (Docker Compose 기반)
# 로컬에 Go가 설치되어 있지 않아도 Docker를 통해 실행 가능합니다.

MIGRATIONS_DIR="./migrations"

# 1. 환경 체크 (Docker 실행 가능 여부)
if ! command -v docker &> /dev/null; then
    echo "❌ 에러: Docker가 설치되어 있지 않습니다."
    exit 1
fi

echo "🚀 모든 DDL 파일에 대한 리스크 분석을 시작합니다..."
echo "📂 검색 경로: $MIGRATIONS_DIR"

# DDL 파일 검색 및 분석
count=0
for file in "$MIGRATIONS_DIR"/*.sql; do
    # 파일이 실제로 존재하지 않으면 스킵
    [ -e "$file" ] || continue

    fileName=$(basename "$file")
    echo ""
    echo "================================================================================"
    echo "🔍 분석 중: $fileName"
    echo "================================================================================"
    
    # Docker Compose를 사용하여 컨테이너 내부의 migraguard 실행
    # --rm: 실행 후 컨테이너 자동 삭제
    # analyze: docker-compose.yml에 정의된 서비스 이름
    # "/app/code/migrations/$fileName": 컨테이너 내부의 파일 경로
    docker compose run --rm analyze analyze "/app/code/migrations/$fileName"
    
    count=$((count + 1))
done

echo ""
echo "✅ 분석 완료! 총 $count 개의 파일을 처리했습니다."
