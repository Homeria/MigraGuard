# Level 5: 어노테이팅 코드 청사진 (Annotated Implementation Blueprint)

이 문서는 MigraGuard의 핵심 로직을 실제 Go 소스 코드와 주석을 결합하여 도면화한 것입니다. 설계자가 의도한 도메인 흐름과 구현의 세부사항을 한눈에 확인할 수 있도록 가공되었습니다.

---

## 🏗️ 1. 분석 및 예측 도메인 (Analysis & Forecast)
`analyze` 명령어 실행 시 호출되는 핵심 오케스트레이터입니다.

### [A] AnalyzeService.Run (Service Layer)
`pkg/migraguard/internal/app/analyze_service.go`

```go
// Run은 DDL 분석의 전체 라이프사이클을 관리합니다.
// 1. SQL 파싱 -> 2. 실시간 리스크 분석 -> 3. 24시간 예측 -> 4. 결과 시각화
func (s *AnalyzeService) Run(ctx context.Context, task AnalysisTask) (*types.AnalysisResponse, error) {
    // [AST Parsing] pg_query_go를 사용하여 SQL을 구문 분석하고 대상 테이블 및 DDL 유형을 식별합니다.
    results, err := analyzer.ParseSQL(string(sqlContent))
    if err != nil { return nil, err }

    for _, res := range results {
        // [Step 1: 실시간 분석] 현재 시점의 DB 메트릭을 기반으로 리스크를 즉시 평가합니다.
        // 리스크 점수와 레벨(Safe/Warning/Danger)이 여기서 결정됩니다.
        report, _ := riskEngine.AnalyzeRisk(ctx, res)

        // [Step 2: 예측 분석] --forecast 플래그가 활성화된 경우, 과거 데이터를 기반으로 24시간 미래를 시뮬레이션합니다.
        if task.Forecast && s.sqlite != nil {
            // 과거 7일간의 지표를 시간대별로 평균화한 프로필을 가져옵니다.
            forecastData, _ := s.sqlite.Get24HourTrafficForecast(ctx, res.TableName)
            
            // 24번의 가상 시뮬레이션을 수행하여 '골든 윈도우'를 탐색합니다.
            fReport, _ := riskEngine.AnalyzeForecast(ctx, res, forecastData)
            forecastReports = append(forecastReports, fReport)
        }
    }
    
    // [Visualization Bridge] 예측 결과가 있으면 Python 시각화 도구가 읽을 수 있도록 CSV를 생성합니다.
    if len(forecastReports) > 0 { 
        s.exportForecastCSV(forecastReports) 
    }

    return &types.AnalysisResponse{ Results: results, Reports: finalReports, ForecastReports: forecastReports }, nil
}
```

---

## 🧠 2. 리스크 엔진 도메인 (Risk Engine)
수학적 모델링을 통해 정량적 점수를 산출하는 시스템의 "두뇌"입니다.

### [B] RiskEngine.AnalyzeForecast (Domain Logic)
`pkg/migraguard/internal/analyzer/risk_calculator.go`

```go
// AnalyzeForecast는 24회 반복 시뮬레이션을 통해 최적의 배포 시간(Golden Window)을 탐색합니다.
func (e *RiskEngine) AnalyzeForecast(ctx context.Context, analysis types.AnalysisResult, forecast []types.ForecastTimeSlot) (*types.ForecastReport, error) {
    
    minScore := 9999.0
    minTPS := 999999.0

    for _, slot := range forecast {
        // [Virtual Snapshot] 해당 시간대의 예상 TPS와 P99를 주입하여 가상 환경을 구축합니다.
        virtualMetrics := types.TableDynamicMetrics{ 
            TPS: slot.ExpectedTPS, 
            P99Time: slot.ExpectedP99 
        }

        // [Strategy Pattern] 5단계 평가 모델을 순차적으로 실행합니다.
        // T_ddl -> T_block -> C_peak -> T_rec -> RiskScore
        for _, evaluator := range e.evaluators {
            evaluator.Evaluate(ctx, analysis, virtualMetrics, tempReport, e.constants)
        }

        // [V3.8 Golden Window Selection Logic] 최적의 시간을 결정하는 지능형 알고리즘
        // 1순위: 리스크 점수가 가장 낮은 시간 (가장 안전한 시간)
        // 2순위: 점수가 같다면 TPS(물리적 부하)가 가장 낮은 시간 (Tie-breaker 적용)
        isLowerScore := slot.RiskScore < (minScore - 0.001)
        isEqualScore := math.Abs(slot.RiskScore-minScore) < 0.001
        isLowerTPS := slot.ExpectedTPS < minTPS

        if isLowerScore || (isEqualScore && isLowerTPS) {
            minScore = slot.RiskScore
            minTPS = slot.ExpectedTPS
            report.BestHour = slot.Hour // 이 시간이 '골든 윈도우'로 추천됨
        }
        
        report.Timeline = append(report.Timeline, slot)
    }
    return report, nil
}
```

---

## 🧪 3. 시뮬레이션 샌드박스 도메인 (Sandbox Engine)
연구 및 검증을 위해 현실적인 가상 데이터를 생성합니다.

### [C] SandboxEngine.SeedScenario (Infra Layer)
`pkg/migraguard/internal/infra/sqlite/sqlite_sandbox.go`

```go
// SeedScenario는 Sine 파형과 노이즈를 결합하여 7일치 고정밀 시계열 데이터를 생성합니다.
func (e *SandboxEngine) SeedScenario(scenario types.SimulationScenario) error {
    
    for i := 0; i <= totalPoints; i++ {
        // [Workload Profiling] 사인파 기반의 일간/주간 트래픽 주기를 계산합니다.
        // 주말에는 트래픽이 감소하고, 특정 시간에는 피크가 발생하는 현실적인 패턴입니다.
        tps := e.profiler.CalculateTPS(t, scenario.History)

        // [Metric Correlation] TPS에 비례하여 P99와 커넥션 수가 증가하도록 수학적으로 모델링합니다.
        // 트래픽이 임계치에 도달할수록 지연 시간이 지수적으로 증가하는 현상을 반영합니다.
        p99 := scenario.PGState.P99TimeMS * math.Exp(tps/scenario.History.PeakTPS-1.0)
        
        // [I/O Simulation] 트래픽 증가 시 시스템 리소스 경합으로 인해 
        // 데이터베이스 캐시 히트율이 점진적으로 하락하는 현상을 시뮬레이션합니다.
        hitRatio := 0.98 - 0.05*(tps/scenario.History.PeakTPS)

        // [Persistence] 생성된 정밀 지표 데이터를 SQLite의 table_metrics 테이블에 적재합니다.
        metricStmt.Exec(t, tableName, currentSize, scenario.PGState.ReplicationLagS, conns, p99, tableTPS, hitBlocks, readBlocks)
    }
    return tx.Commit()
}
```
