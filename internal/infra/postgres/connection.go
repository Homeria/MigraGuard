package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresConfig는 데이터베이스 연결에 필요한 설정 정보를 보관합니다.
type PostgresConfig struct {
	URL string // 접속 URL (예: postgres://user:pass@host:5432/db)
}

// ConnectPostgres는 지정된 URL을 사용하여 PostgreSQL에 대한 물리적인 커넥션 풀을 생성합니다.
// Args:
//   - ctx: 애플리케이션 컨텍스트
//   - url: Postgres 접속 문자열
// Returns:
//   - *pgxpool.Pool: 활성화된 커넥션 풀 객체
//   - error: 연결 설정 오류 또는 네트워크 연결 실패 시 반환
func ConnectPostgres(ctx context.Context, url string) (*pgxpool.Pool, error) {
	// 1. 접속 문자열의 형식을 검증하고 설정을 파싱합니다.
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("Postgres 연결 설정 파싱 실패: %w", err)
	}

	// 2. 파싱된 설정을 바탕으로 새로운 커넥션 풀 객체를 생성합니다.
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("Postgres 연결 풀 생성 실패: %w", err)
	}

	// 3. 실제 데이터베이스 서버와 통신이 가능한지 Ping 테스트를 수행합니다.
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Postgres 가용성 확인(Ping) 실패: %w", err)
	}

	return pool, nil
}
