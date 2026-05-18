# Level 4: 함수 단위 마이크로 플로우 (Micro-Flows)

이 다이어그램은 MigraGuard의 핵심 로직을 함수 단위로 쪼개어, 각 단계가 어떤 파일의 어떤 함수에서 어떤 인자를 가지고 수행되는지 상세히 묘사합니다.

---

## 1. `analyze` 명령어 전체 실행 로직

```mermaid
flowchart TD
    Step1["<b>File:</b> analyze_service.go<br/><b>Func:</b> Run()<br/><b>Args:</b> ctx, task<br/><b>Role:</b> DDL 분석 전체 프로세스 오케스트레이션"]
    
    Step2["<b>File:</b> sql_parser.go<br/><b>Func:</b> ParseSQL()<br/><b>Args:</b> sqlContent<br/><b>Role:</b> SQL을 AST로 변환 및 DDL 유형(Table/Index/Column) 식별"]
    
    Step3["<b>File:</b> risk_calculator.go<br/><b>Func:</b> AnalyzeRisk()<br/><b>Args:</b> ctx, analysisResult<br/><b>Role:</b> 현재 DB 상태 기준 5단계 리스크 점수 산출"]
    
    Step4["<b>File:</b> sqlite_analyzer.go<br/><b>Func:</b> Get24HourTrafficForecast()<br/><b>Args:</b> ctx, tableName<br/><b>Role:</b> 과거 7일 지표를 기반으로 24시간 트래픽 프로필 생성"]
    
    Step5["<b>File:</b> risk_calculator.go<br/><b>Func:</b> AnalyzeForecast()<br/><b>Args:</b> ctx, analysis, forecastData<br/><b>Role:</b> 24회 반복 시뮬레이션 및 최적의 골든 윈도우 탐색"]
    
    Step6["<b>File:</b> analyze_service.go<br/><b>Func:</b> exportForecastCSV()<br/><b>Args:</b> forecastReports<br/><b>Role:</b> 시각화 도구(Python)용 CSV 데이터 익스포트"]

    Step1 --> Step2
    Step2 --> Step3
    Step3 --> Step4
    Step4 --> Step5
    Step5 --> Step6
    
    style Step1 fill:#f9f,stroke:#333,stroke-width:2px
    style Step3 fill:#fff3e0,stroke:#e65100,stroke-width:2px
    style Step5 fill:#fff3e0,stroke:#e65100,stroke-width:2px
```

---

## 2. 예측 엔진 및 Tie-Breaker 내부 로직

```mermaid
flowchart TD
    F1["<b>File:</b> risk_calculator.go<br/><b>Func:</b> AnalyzeForecast()<br/><b>Args:</b> ctx, analysis, forecast<br/><b>Role:</b> 24시간 가상 시뮬레이션 루프 시작"]
    
    F2["<b>File:</b> risk_evaluator.go<br/><b>Func:</b> Evaluate()<br/><b>Args:</b> ctx, metrics, report, constants<br/><b>Role:</b> 가상 지표(ExpectedTPS/P99) 주입 후 5단계 모델 실행"]
    
    F3["<b>File:</b> risk_calculator.go<br/><b>내부 로직</b><br/><b>Logic:</b> math.Abs(score - minScore) < 0.001<br/><b>Role:</b> 동점 리스크 점수 발생 시 TPS Tie-breaker 적용"]
    
    F4["<b>File:</b> risk_calculator.go<br/><b>Func:</b> report.BestHour update<br/><b>Args:</b> currentHour<br/><b>Role:</b> 가장 낮은 리스크 & TPS를 가진 골든 윈도우 최종 확정"]

    F1 --> F2
    F2 --> F3
    F3 --> F4
    
    style F2 fill:#e1f5fe,stroke:#01579b
    style F3 fill:#ffecb3,stroke:#ff6f00,stroke-width:3px
```

---

## 3. 샌드박스 데이터 생성 및 시딩 로직

```mermaid
flowchart TD
    S1["<b>File:</b> simulate_service.go<br/><b>Func:</b> Run()<br/><b>Args:</b> ctx, scenarioPath, force<br/><b>Role:</b> 시나리오 YAML 로드 및 샌드박스 DB 초기화"]
    
    S2["<b>File:</b> sqlite_sandbox.go<br/><b>Func:</b> SeedScenario()<br/><b>Args:</b> scenario<br/><b>Role:</b> 7일치(10,080분) 타임시리즈 데이터 생성 루프"]
    
    S3["<b>File:</b> sqlite_sandbox.go<br/><b>Func:</b> CalculateTPS()<br/><b>Args:</b> time, profile<br/><b>Role:</b> Sine Wave + Weekly + Noise 기반 트래픽 산출"]
    
    S4["<b>File:</b> sqlite_sandbox.go<br/><b>데이터베이스 액션</b><br/><b>Args:</b> metricStmt.Exec(...)<br/><b>Role:</b> 생성된 가상 지표를 SQLite에 Batch Insert"]

    S1 --> S2
    S2 --> S3
    S3 --> S4
    
    style S3 fill:#f1f8e9,stroke:#33691e
```
