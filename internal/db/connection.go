package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAdapter manages connections to the target PostgreSQL database.
// 대상 PostgreSQL 데이터베이스에 대한 연결을 관리합니다.
type PostgresAdapter struct {
	pool *pgxpool.Pool
	url  string
}

// Config represents the database connection configuration.
// 데이터베이스 연결 설정을 나타냅니다.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewPostgresAdapter creates a new PostgresAdapter instance.
// 새로운 PostgresAdapter 인스턴스를 생성합니다.
func NewPostgresAdapter(url string) *PostgresAdapter {
	return &PostgresAdapter{
		url: url,
	}
}

// Connect establishes a connection pool to the PostgreSQL server.
// PostgreSQL 서버에 대한 연결 풀을 생성합니다.
func (a *PostgresAdapter) Connect(ctx context.Context) error {

	// 현재 설정값 불러오기
	config, err := pgxpool.ParseConfig(a.url)
	if err != nil {
		return fmt.Errorf("failed to parse connection URL: %w", err)
	}

	// TODO: Configurable connection pool settings (migraguard.yaml)

	// MaxConns : MigraGuard에서 동시에 DB에 던질 수 있는 최대 연결(세션) 수
	config.MaxConns = 10

	// MinConns : 연결 풀에서 유지할 최소 연결 수
	config.MinConns = 2

	// MaxConnLifetime : 연결 풀에서 개별 연결이 유지될 수 있는 최대 시간
	config.MaxConnLifetime = time.Hour

	// 수정한 config 값을 시스템에 주입하여 connection pool 생성
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 연결 확인을 위해 DB에 핑 보내기
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	a.pool = pool
	return nil
}

// Close terminates the connection pool.
// 연결 풀을 종료합니다.
func (a *PostgresAdapter) Close() {
	if a.pool != nil {
		a.pool.Close()
	}
}

// GetPool returns the underlying pgxpool.Pool instance.
// 내부의 pgxpool.Pool 인스턴스를 반환합니다.
func (a *PostgresAdapter) GetPool() *pgxpool.Pool {
	return a.pool
}
