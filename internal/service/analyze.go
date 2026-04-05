package service

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/engine"
	"github.com/Homeria/MigraGuard/internal/errors"
	"github.com/Homeria/MigraGuard/internal/parser"
)

// AnalyzeService handles the end-to-end migration analysis pipeline.
type AnalyzeService struct {
	pg        db.PostgresClient
	sqlite    db.SQLiteClient
	constants engine.RiskConstants
	Verbose   bool
}

// NewAnalyzeService creates a new instance of AnalyzeService.
func NewAnalyzeService(pg db.PostgresClient, sqlite db.SQLiteClient, constants engine.RiskConstants, verbose bool) *AnalyzeService {
	return &AnalyzeService{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
		Verbose:   verbose,
	}
}

// AnalysisTask represents the input for an analysis operation.
type AnalysisTask struct {
	SQLPath string
}

// AnalysisResponse holds the results of the analysis for the reporter.
type AnalysisResponse struct {
	Results []parser.AnalysisResult
	Reports []*engine.RiskAnalysisReport
}

// Run executes the full analysis pipeline: Load -> Parse -> Validate -> Analyze.
func (s *AnalyzeService) Run(ctx context.Context, task AnalysisTask) (*AnalysisResponse, error) {

	// 1. Load SQL file
	sqlContent, err := os.ReadFile(task.SQLPath)
	if err != nil {
		return nil, errors.Wrap(err, "AnalyzeService.Run", "failed to read SQL file")
	}

	// 2. Static Analysis (AST Parsing)
	// internal/parser/ast.go - 서비스 생성 시 파라미터로 받은 SQL 파일 경로를 기반으로 SQL 파일을 읽어 AST로 파싱하여 반환
	results, err := parser.ParseSQL(string(sqlContent))
	if err != nil {
		return nil, errors.Wrap(err, "AnalyzeService.Run", "SQL parser error")
	}

	// SQL 파일에서 유효한 DDL 분석 결과가 없는 경우
	if len(results) == 0 {
		return nil, errors.ErrInvalidSQL
	}

	// 3. Initialize Risk Engine
	// RiskEngine 인스턴스 생성 - 타겟 DB 어댑터, SQLite 어댑터, 위험 분석에 필요한 상수값을 주입하여 초기화
	riskEngine := engine.NewRiskEngine(s.pg, s.sqlite, s.constants)
	// Verbose 필드 값을 RiskEngine에 전달하여 디버그 로그 출력 여부 결정
	riskEngine.Verbose = s.Verbose

	var finalReports []*engine.RiskAnalysisReport
	var validResults []parser.AnalysisResult

	// 4. Validation & Analysis Loop

	for _, res := range results {
		// [L61] Schema Validation
		// internal/db/workload.go - ValidateSchema 메서드를 호출하여 분석 결과에 포함된 테이블과 컬럼이 타겟 DB에 존재하는지 검증
		if err := s.pg.ValidateSchema(ctx, res.TableName, res.Columns); err != nil {
			if errors.Is(err, errors.ErrTableNotFound) || errors.Is(err, errors.ErrColumnNotFound) {
				if s.Verbose {
					fmt.Printf("[DEBUG] Validation failed for table '%s': %v. Skipping.\n", res.TableName, err)
				}
				continue
			}
			return nil, errors.WrapWithTable(err, "AnalyzeService.Run", res.TableName, "schema validation error")
		}

		// [L31~L35] Risk Analysis
		// 생성한 RiskEngine의 AnalyzeRisk 메서드를 호출하여 각 유효한 분석 결과에 대해 위험 분석을 수행
		// 분석 결과를 RiskAnalysisReport 구조체로 반환
		report, err := riskEngine.AnalyzeRisk(ctx, res)
		if err != nil {
			return nil, errors.WrapWithTable(err, "AnalyzeService.Run", res.TableName, "risk analysis execution failed")
		}

		// 유효한 분석 결과와 해당 결과에 대한 위험 분석 보고서를 각각 validResults와 finalReports 슬라이스에 저장
		validResults = append(validResults, res)
		finalReports = append(finalReports, report)
	}

	// 5. Finalize Response

	// 유효한 분석 결과가 없는 경우, 테이블을 찾을 수 없다는 오류 반환
	if len(validResults) == 0 {
		return nil, errors.Wrap(errors.ErrTableNotFound, "AnalyzeService.Run", "no valid tables found for analysis")
	}

	// 유효한 분석 결과와 해당 결과에 대한 위험 분석 보고서를 포함하는 AnalysisResponse 구조체를 반환
	return &AnalysisResponse{
		Results: validResults,
		Reports: finalReports,
	}, nil
}
