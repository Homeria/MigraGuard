package app

import (
	"context"
	"errors"
	"os"

	"github.com/Homeria/MigraGuard/internal/analyzer"
	migraErrors "github.com/Homeria/MigraGuard/internal/shared/errors"
	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// AnalyzeService는 마이그레이션 분석 파이프라인의 전체 실행을 조율하는 서비스입니다.
type AnalyzeService struct {
	pg        types.PostgresClient
	sqlite    types.SQLiteClient
	constants analyzer.RiskConstants
	Verbose   bool
}

// NewAnalyzeService는 AnalyzeService 인스턴스를 생성하고 초기화합니다.
//
// Args:
//   - pg: PostgresClient 구현체
//   - sqlite: SQLiteClient 구현체
//   - constants: 리스크 임계치
//   - verbose: 로그 상세 출력 옵션
//
// Returns:
//   - *AnalyzeService: 초기화된 서비스 객체
func NewAnalyzeService(pg types.PostgresClient, sqlite types.SQLiteClient, constants analyzer.RiskConstants, verbose bool) *AnalyzeService {
	// 1. 필드 주입 및 서비스 조립
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

// Run은 SQL 로드부터 최종 리스크 산출까지의 분석 워크플로우를 가동합니다.
//
// Args:
//   - ctx: 어플리케이션 컨텍스트
//   - task: 분석 대상 파일 정보
//
// Returns:
//   - *AnalysisResponse: 최종 리포트 집합
//   - error: 파이프라인 에러
func (s *AnalyzeService) Run(ctx context.Context, task AnalysisTask) (*AnalysisResponse, error) {
	// 1. 파일 시스템 로드: SQL 내용 읽기
	sqlContent, err := os.ReadFile(task.SQLPath)
	if err != nil {
		return nil, migraErrors.Wrap(err, "AnalyzeService.Run", "파일 로드 실패")
	}

	// 2. 정적 분석: DDL 특성 및 테이블 추출
	results, err := analyzer.ParseSQL(string(sqlContent))
	if err != nil {
		return nil, migraErrors.Wrap(err, "AnalyzeService.Run", "파싱 오류")
	}
	if len(results) == 0 {
		return nil, migraErrors.ErrInvalidSQL
	}

	// 3. 엔진 초기화 및 변수 할당
	riskEngine := analyzer.NewRiskEngine(s.pg, s.sqlite, s.constants)
	riskEngine.Verbose = s.Verbose
	var finalReports []*analyzer.RiskAnalysisReport
	var validResults []analyzer.AnalysisResult

	// 4. 분석 루프: 스키마 존재 확인 및 5단계 리스크 모델 적용
	for _, res := range results {
		// 4-1. 실효성 검증: 타겟 DB에 실제 스키마 존재 여부 확인
		if err := s.pg.CheckTableSchemaPresence(ctx, res.TableName, res.Columns); err != nil {
			if errors.Is(err, migraErrors.ErrTableNotFound) || errors.Is(err, migraErrors.ErrColumnNotFound) {
				continue
			}
			return nil, migraErrors.WrapWithTable(err, "AnalyzeService.Run", res.TableName, "검증 실패")
		}

		// 4-2. 리스크 엔진 기동: 정량적 점수 산출
		report, err := riskEngine.AnalyzeRisk(ctx, res)
		if err != nil {
			return nil, migraErrors.WrapWithTable(err, "AnalyzeService.Run", res.TableName, "분석 실패")
		}

		validResults = append(validResults, res)
		finalReports = append(finalReports, report)
	}

	// 5. 결과 조합: 분석 완료 데이터 반환
	if len(validResults) == 0 {
		return nil, migraErrors.Wrap(migraErrors.ErrTableNotFound, "AnalyzeService.Run", "유효 대상 없음")
	}

	return &AnalysisResponse{
		Results: validResults,
		Reports: finalReports,
	}, nil
}
