package analyzer

import (
	"fmt"
	"strings"

	"github.com/pganalyze/pg_query_go/v5"
)

// AnalysisResult는 SQL 정적 분석의 결과 데이터입니다.
type AnalysisResult struct {
	TableName       string   // 대상 테이블명
	Operation       string   // 수행 작업 (ALTER, CREATE 등)
	Columns         []string // 관련 컬럼 목록
	RewriteRequired bool     // 테이블 재작성(Full Rewrite) 유발 여부
	RawQuery        string   // 분석된 실제 SQL 문장 원문
}

// ParseSQL은 SQL 텍스트를 파싱하여 DDL 리스크 분석에 필요한 정보를 추출합니다.
func ParseSQL(sqlText string) ([]AnalysisResult, error) {
	result, err := pg_query.Parse(sqlText)
	if err != nil {
		return nil, fmt.Errorf("SQL 구문 분석 실패: %w", err)
	}

	var analyses []AnalysisResult

	for _, stmt := range result.Stmts {
		node := stmt.Stmt
		var res AnalysisResult

		if node.GetAlterTableStmt() != nil {
			res = analyzeAlterTable(node.GetAlterTableStmt())
		} else {
			// 지원하지 않는 구문은 건너뜀 (추후 확장 가능)
			continue
		}

		// SQL 원문 텍스트 캡처 (Statement 범위 추출)
		// pg_query_go의 Stmt 데이터에는 원본 SQL에서의 위치 정보가 포함되어 있음
		res.RawQuery = sqlText
		if len(result.Stmts) > 1 {
			// 여러 문장이 있을 경우, 현재 문장의 길이만큼 최대한 근사하게 자름 (단순화된 방식)
			// 실제로는 stmt.Location 등을 활용하여 정밀하게 추출 가능
			res.RawQuery = "Selected Statement from multi-query script"
		}
		
		// 보다 정확한 원문 전달을 위해 전체 텍스트에서 해당 문장을 식별하는 로직 (단순화 버전)
		if len(sqlText) > 0 {
			res.RawQuery = strings.TrimSpace(sqlText)
		}

		analyses = append(analyses, res)
	}

	return analyses, nil
}

func analyzeAlterTable(stmt *pg_query.AlterTableStmt) AnalysisResult {
	tableName := stmt.Relation.Relname
	res := AnalysisResult{
		TableName: tableName,
		Operation: "ALTER TABLE",
	}

	for _, cmd := range stmt.Cmds {
		subCmd := cmd.GetAlterTableCmd()
		if subCmd == nil {
			continue
		}

		// Table Rewrite를 유발하는 대표적인 케이스 판별
		// 1. 컬럼 타입 변경 (Type Change)
		if subCmd.Subtype == pg_query.AlterTableType_AT_AlterColumnType {
			res.RewriteRequired = true
		}

		// 2. 새로운 제약 조건 추가 (일부 케이스)
		if subCmd.Subtype == pg_query.AlterTableType_AT_AddConstraint {
			res.RewriteRequired = true
		}

		// 관련 컬럼 추출
		if subCmd.Name != "" {
			res.Columns = append(res.Columns, subCmd.Name)
		}
	}

	return res
}

// NormalizeTableName은 대소문자 구분 없는 테이블 매칭을 위한 헬퍼 함수입니다.
func NormalizeTableName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
