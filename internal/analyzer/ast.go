package analyzer

import (
	"fmt"
	"strings"

	"github.com/pganalyze/pg_query_go/v5"
)

// AnalysisResult는 SQL 정적 분석을 통해 도출된 결과 데이터를 담는 구조체입니다.
type AnalysisResult struct {
	TableName       string   // 대상 테이블명
	Operation       string   // SQL 작업 유형
	Columns         []string // 관련 컬럼 목록
	RewriteRequired bool     // Table Rewrite 유발 여부
	RawQuery        string   // 원본 SQL 문장
}

// ParseSQL은 입력받은 SQL 문자열을 AST로 파싱하여 리스크 분석 정보를 추출합니다.
//
// Args:
//   - sqlText: 원시 SQL 문자열
//
// Returns:
//   - []AnalysisResult: 분석 결과 슬라이스
//   - error: 파싱 실패 에러
func ParseSQL(sqlText string) ([]AnalysisResult, error) {
	// 1. Postgres 파서 기동
	result, err := pg_query.Parse(sqlText)
	if err != nil {
		return nil, fmt.Errorf("SQL 구문 분석 실패: %w", err)
	}

	var analyses []AnalysisResult

	// 2. 문장별 순회 및 DDL 특징 추출
	for _, stmt := range result.Stmts {
		node := stmt.Stmt
		var res AnalysisResult

		if node.GetAlterTableStmt() != nil {
			res = analyzeAlterTable(node.GetAlterTableStmt())
		} else {
			continue
		}

		// 3. 원문 텍스트 매핑
		if len(sqlText) > 0 {
			res.RawQuery = strings.TrimSpace(sqlText)
		}

		analyses = append(analyses, res)
	}

	return analyses, nil
}

// analyzeAlterTable은 ALTER TABLE 노드를 분석하여 재작성 발생 여부를 판별합니다.
//
// Args:
//   - stmt: 파싱된 ALTER 문장 노드
//
// Returns:
//   - AnalysisResult: 상세 분석 정보
func analyzeAlterTable(stmt *pg_query.AlterTableStmt) AnalysisResult {
	// 1. 기본 정보 할당
	tableName := stmt.Relation.Relname
	res := AnalysisResult{
		TableName: tableName,
		Operation: "ALTER TABLE",
	}

	// 2. 서브 명령 순회 및 고위험 패턴(Type Change, Constraint) 감지
	for _, cmd := range stmt.Cmds {
		subCmd := cmd.GetAlterTableCmd()
		if subCmd == nil { continue }

		if subCmd.Subtype == pg_query.AlterTableType_AT_AlterColumnType {
			res.RewriteRequired = true
		}
		if subCmd.Subtype == pg_query.AlterTableType_AT_AddConstraint {
			res.RewriteRequired = true
		}

		// 3. 대상 컬럼 수집
		if subCmd.Name != "" {
			res.Columns = append(res.Columns, subCmd.Name)
		}
	}

	return res
}

// NormalizeTableName은 일관된 처리를 위해 테이블명을 소문자로 정규화합니다.
//
// Args:
//   - name: 원본 이름
//
// Returns:
//   - string: 정규화된 이름
func NormalizeTableName(name string) string {
	// 1. 공백 제거 및 소문자 변환
	return strings.ToLower(strings.TrimSpace(name))
}
