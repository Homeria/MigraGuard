package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/analyzer"
	migraErrors "github.com/Homeria/MigraGuard/internal/shared/errors"
	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// AnalyzeService는 마이그레이션 리스크 분석의 전체 오케스트레이션을 담당하는 서비스입니다.
type AnalyzeService struct {
	pg        types.PostgresClient
	sqlite    types.SQLiteClient
	constants analyzer.RiskConstants
	Verbose   bool
}

// NewAnalyzeService는 AnalyzeService의 새로운 인스턴스를 생성합니다.
func NewAnalyzeService(pg types.PostgresClient, sqlite types.SQLiteClient, constants analyzer.RiskConstants, verbose bool) *AnalyzeService {
	return &AnalyzeService{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
		Verbose:   verbose,
	}
}

// AnalysisTask는 분석 작업에 필요한 입력 파일 경로를 정의합니다.
type AnalysisTask struct {
	SQLPath string
}

// AnalysisResponse는 리포터(Reporter)에게 전달될 최종 분석 결과 번들입니다.
type AnalysisResponse struct {
	Results []analyzer.AnalysisResult
	Reports []*analyzer.RiskAnalysisReport
}

// Run은 SQL 로드 -> 파싱 -> 스키마 검증 -> 리스크 엔진 분석으로 이어지는 전체 파이프라인을 실행합니다.
func (s *AnalyzeService) Run(ctx context.Context, task AnalysisTask) (*AnalysisResponse, error) {

	// 1. SQL 파일 읽기
	sqlContent, err := os.ReadFile(task.SQLPath)
	if err != nil {
		return nil, migraErrors.Wrap(err, "AnalyzeService.Run", "지정된 SQL 파일을 읽을 수 없습니다.")
	}

	// 2. 정적 분석 (AST Parsing)
	results, err := analyzer.ParseSQL(string(sqlContent))
	if err != nil {
		return nil, migraErrors.Wrap(err, "AnalyzeService.Run", "SQL 파서 가동 중 오류가 발생했습니다.")
	}

	if len(results) == 0 {
		return nil, migraErrors.ErrInvalidSQL
	}

	// 3. 리스크 엔진 초기화
	riskEngine := analyzer.NewRiskEngine(s.pg, s.sqlite, s.constants)
	riskEngine.Verbose = s.Verbose

	var finalReports []*analyzer.RiskAnalysisReport
	var validResults []analyzer.AnalysisResult

	// 4. 스키마 유효성 검증 및 리스크 분석 수행 루프
	for _, res := range results {
		// 운영 DB에 실제 테이블/컬럼이 존재하는지 사전 체크
		if err := s.pg.CheckTableSchemaPresence(ctx, res.TableName, res.Columns); err != nil {
			if errors.Is(err, migraErrors.ErrTableNotFound) || errors.Is(err, migraErrors.ErrColumnNotFound) {
				if s.Verbose {
					fmt.Printf("[DEBUG] 검증 실패: 테이블 '%s'가 존재하지 않거나 스키마가 불일치하여 분석을 건너뜁니다.\n", res.TableName)
				}
				continue
			}
			return nil, migraErrors.WrapWithTable(err, "AnalyzeService.Run", res.TableName, "스키마 검증 도중 오류가 발생했습니다.")
		}

		// 정밀 리스크 모델 실행
		report, err := riskEngine.AnalyzeRisk(ctx, res)
		if err != nil {
			return nil, migraErrors.WrapWithTable(err, "AnalyzeService.Run", res.TableName, "리스크 분석 엔진 실행 실패")
		}

		validResults = append(validResults, res)
		finalReports = append(finalReports, report)
	}

	// 분석 가능한 유효 결과가 하나도 없는 경우 에러 처리
	if len(validResults) == 0 {
		return nil, migraErrors.Wrap(migraErrors.ErrTableNotFound, "AnalyzeService.Run", "분석을 진행할 수 있는 유효한 대상 테이블을 찾지 못했습니다.")
	}

	return &AnalysisResponse{
		Results: validResults,
		Reports: finalReports,
	}, nil
}
