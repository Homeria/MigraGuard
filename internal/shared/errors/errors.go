package errors

import (
	"errors"
	"fmt"
)

// 프로젝트 전반에서 공통으로 사용되는 커스텀 에러 정의
var (
	ErrDatabaseConn  = errors.New("데이터베이스 연결에 실패했습니다")
	ErrTableNotFound = errors.New("대상 테이블을 찾을 수 없습니다")
	ErrColumnNotFound = errors.New("대상 컬럼을 찾을 수 없습니다")
	ErrInvalidSQL    = errors.New("유효하지 않은 SQL 구문입니다")
	ErrAnalysis      = errors.New("분석 엔진 실행 도중 오류가 발생했습니다")
)

// Wrap은 에러에 문맥(Context) 정보를 추가하여 새로운 에러를 생성합니다.
func Wrap(err error, op string, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("[%s] %s: %w", op, message, err)
}

// WrapWithTable은 테이블 명시가 필요한 에러 상황에서 문맥 정보를 추가합니다.
func WrapWithTable(err error, op string, table string, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("[%s] 테이블(%s) - %s: %w", op, table, message, err)
}
