package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresConfig는 PostgreSQL 데이터베이스 연결 정보를 담는 구조체입니다.
type PostgresConfig struct {
	URL string
}

// ConnectPostgres는 제공된 접속 문자열을 사용하여 PostgreSQL 연결 풀을 생성하고 상태를 확인(Ping)합니다.
func ConnectPostgres(ctx context.Context, url string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("Postgres 연결 설정 파싱 실패: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("Postgres 연결 풀 생성 실패: %w", err)
	}

	// 실제 DB와 통신이 가능한지 즉시 확인
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Postgres 가용성 확인(Ping) 실패: %w", err)
	}

	return pool, nil
}
