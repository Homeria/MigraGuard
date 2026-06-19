# MigraGuard 발표 시연 가이드

이 문서는 캡스톤디자인 발표에서 MigraGuard의 핵심 동작을 약 5분 안에 시연하기 위한 진행 순서이다.

시연의 핵심 주장은 다음과 같다.

> 정적 린터를 통과한 동일한 DDL도 실행 시점의 작업 부하와 서버 처리 여력에 따라 MigraGuard에서 서로 다른 위험 등급으로 판단될 수 있다.

## 1. 시연 범위

시연에서는 실제 운영 DB에 위험한 DDL을 실행하지 않는다. YAML 시나리오로 재현 가능한 작업 부하 데이터를 생성하고, 해당 데이터와 DDL의 정적 특성을 결합하여 위험도를 분석한다.

시연 순서는 다음과 같다.

1. 대표 DDL 확인
2. Squawk 정적 검사 통과 확인
3. 대규모 로그 적재 시나리오 생성
4. Default 설정에서 MigraGuard Danger 판정 확인
5. High Capacity 설정에서 Safe 판정 확인
6. 동일한 DDL도 실행 환경에 따라 위험도가 달라진다는 결론 설명

## 2. 발표 전 준비

프로젝트 루트에서 PowerShell 또는 CMD를 실행한다.

```powershell
cd C:\Users\CGH\Documents\GitHub\MigraGuard
```

실행 파일이 없거나 소스가 변경되었다면 다시 빌드한다.

```powershell
go build -o migraguard.exe ./cmd/migraguard
```

명령어 도움말과 실행 파일을 확인한다.

```powershell
.\migraguard.exe --help
```

발표 전에 아래 전체 시연 명령을 한 번 실행해 둔다. 원격 접속을 사용할 경우 데스크톱의 절전 모드를 해제하고, 노트북에는 이 문서와 주요 결과 화면을 별도로 열어 둔다.

## 3. 5분 시연 대본

### 3.1 대표 DDL 확인

```powershell
Get-Content experiments\ddl\021_comparison_linter_pass_dynamic_danger.sql
```

설명할 내용:

- `order_event_logs.created_at`에 인덱스를 생성한다.
- `CONCURRENTLY`로 쓰기 차단을 줄였다.
- `IF NOT EXISTS`로 재실행 안전성을 확보했다.
- 잠금 및 문장 실행 시간 제한을 설정했다.

### 3.2 Squawk 정적 검사

```powershell
cmd /c npx --yes squawk-cli@2.58.0 experiments\ddl\021_comparison_linter_pass_dynamic_danger.sql
```

확인할 결과:

```text
Found 0 issues in 1 file
```

발표 멘트:

> Squawk의 정적 검사에서는 이 DDL에 규칙 위반이 없으므로 배포 전 검사를 통과합니다. 하지만 이 결과만으로 현재 데이터베이스가 해당 작업을 감당할 수 있는지는 판단할 수 없습니다.

### 3.3 작업 부하 시나리오 생성

먼저 시나리오 내용을 간단히 확인한다.

```powershell
Get-Content experiments\scenarios\06_heavy_log_insert.yaml
```

설명할 내용:

- 대상 테이블: `order_event_logs`
- 테이블 크기: 약 100GB
- 활성 커넥션: 100개
- 높은 쓰기 요청과 주기적인 배치 작업
- 복제 지연: 1.2초

시뮬레이션 데이터를 생성한다.

```powershell
.\migraguard.exe simulate --scenario experiments\scenarios\06_heavy_log_insert.yaml --force
```

생성되는 샌드박스 파일:

```text
exp_06_heavy_log_insert.db
```

발표 멘트:

> 실제 운영 DB에서 위험한 DDL을 반복 실행하기 어렵기 때문에, 보고서에서는 YAML 시나리오를 이용해 재현 가능한 작업 부하 데이터를 생성했습니다.

### 3.4 Default 설정 분석

```powershell
.\migraguard.exe analyze experiments\ddl\021_comparison_linter_pass_dynamic_danger.sql --sandbox exp_06_heavy_log_insert.db --config experiments\configs\cases\default.yaml --output console
```

확인할 항목:

- 최종 등급: `Danger`
- `T_ddl`: 예상 DDL 수행 시간
- `T_block`: 예상 차단 시간
- `C_peak`: 차단 시간 동안의 예상 최대 커넥션 수요
- `T_rec`: 누적 요청 처리 후 예상 회복 시간
- `Risk Score`: 최종 위험 점수

발표 멘트:

> 동일한 DDL을 대규모 로그 적재 환경과 결합하면 요청이 대기하는 동안 예상 커넥션 수요가 증가합니다. Default 설정에서는 이 값이 허용 한계에 도달하여 최종적으로 Danger가 산출됩니다.

시뮬레이션에는 일부 변동성이 있으므로 보고서와 세부 숫자가 조금 달라질 수 있다. 최종 등급과 각 지표가 증가하는 흐름을 중심으로 설명한다.

### 3.5 High Capacity 설정 분석

DDL과 작업 부하는 그대로 유지하고 서버 처리 여력만 변경한다.

```powershell
.\migraguard.exe analyze experiments\ddl\021_comparison_linter_pass_dynamic_danger.sql --sandbox exp_06_heavy_log_insert.db --config experiments\configs\cases\highcapacity.yaml --output console
```

확인할 결과:

- 최종 등급: `Safe`
- 최대 허용 커넥션 수와 최대 처리량 증가
- Default보다 낮아진 `Risk Score`

발표 멘트:

> DDL과 작업 부하는 동일하지만 서버 처리 여력이 증가하면 커넥션 압력이 허용 범위 안에 머물러 Safe로 낮아집니다. 이를 통해 MigraGuard의 위험도가 DDL에 고정된 값이 아니라 실행 환경과 결합해 결정된다는 것을 확인할 수 있습니다.

### 3.6 시연 결론

> Squawk는 DDL 자체의 정적 규칙 위반을 검사하고, MigraGuard는 정적 검사를 통과한 DDL을 현재 작업 부하와 서버 처리 여력에서 실행할 때의 영향을 추가로 평가합니다. 두 도구는 대체 관계가 아니라 배포 전에 순차적으로 사용할 수 있는 상호 보완적인 검증 단계입니다.

## 4. 보고서 비교 실험 재현 명령

이 절은 보고서의 그림 4~6을 생성한 비교 실험을 재현할 때 사용한다. 20개 DDL, 20개 작업 부하 시나리오와 여러 서버 설정을 반복 분석하므로 발표 현장에서 전체를 실행하지 않는다. 발표 전 검증이나 질문 대응에 사용한다.

### 4.1 Squawk 통과 DDL 15개 검증

DDL 022~036은 Squawk를 통과하도록 구성한 Safe, Warning, Danger 설계 DDL 각 5개이다.

```powershell
$passDdls = Get-ChildItem .\experiments\ddl\*.sql |
    Where-Object { $_.BaseName -match '^(02[2-9]|03[0-6])_' } |
    Sort-Object Name

& "$env:ProgramFiles\nodejs\npx.cmd" --yes squawk-cli@2.58.0 @($passDdls.FullName)
```

확인할 결과:

```text
Found 0 issues in 15 files
```

### 4.2 Squawk 경고 DDL 5개 검증

DDL 037~041은 비동시 인덱스 생성, 컬럼 타입 변경, NOT NULL 직접 설정, 즉시 검증 제약조건 추가와 컬럼명 변경을 포함한다.

```powershell
$warningDdls = Get-ChildItem .\experiments\ddl\*.sql |
    Where-Object { $_.BaseName -match '^(03[7-9]|04[0-1])_' } |
    Sort-Object Name

& "$env:ProgramFiles\nodejs\npx.cmd" --yes squawk-cli@2.58.0 @($warningDdls.FullName)
```

확인할 결과:

```text
Found 5 issues in 5 files (checked 5 source files)
```

Squawk는 경고가 있으면 종료 코드 1을 반환한다. 이 경우는 명령 실패가 아니라 실험에서 의도한 결과이다.

### 4.3 Default 설정 비교 실험

DDL 022~041과 20개 작업 부하 시나리오를 Default 설정에서 결합하여 총 400개 조합을 분석한다.

```powershell
.\scripts\ps1\analyze\run-ddl-scenario-config-matrix.ps1 `
    -RunName linter_comparison_default_demo `
    -DdlStart 22 `
    -DdlEnd 41 `
    -ConfigNames default
```

주요 결과 파일:

```text
experiments\reports\batch_runs\linter_comparison_default_demo\matrix_results.csv
experiments\reports\batch_runs\linter_comparison_default_demo\ddl_summary.csv
experiments\reports\batch_runs\linter_comparison_default_demo\ddl_by_config_summary.csv
```

Squawk 통과 DDL과 경고 DDL의 히트맵을 생성한다.

```powershell
python .\tools\visualization\analyze\plot_linter_migraguard_heatmaps.py `
    .\experiments\reports\batch_runs\linter_comparison_default_demo\matrix_results.csv `
    --config default
```

생성되는 그림:

```text
experiments\reports\batch_runs\linter_comparison_default_demo\figures\linter_migraguard_comparison\01_squawk_pass_intended_vs_migraguard.png
experiments\reports\batch_runs\linter_comparison_default_demo\figures\linter_migraguard_comparison\02_squawk_warning_migraguard_results.png
```

### 4.4 Low, Default, High Capacity 비교 실험

동일한 20개 DDL과 20개 작업 부하 시나리오에 Low Capacity, Default, High Capacity 설정을 각각 적용하여 총 1,200개 조합을 분석한다.

```powershell
.\scripts\ps1\analyze\run-ddl-scenario-config-matrix.ps1 `
    -RunName capacity_comparison_demo `
    -DdlStart 22 `
    -DdlEnd 41 `
    -ConfigNames lowcapacity,default,highcapacity
```

Config 비교 히트맵을 생성한다.

```powershell
python .\tools\visualization\analyze\plot_capacity_config_heatmaps.py `
    .\experiments\reports\batch_runs\capacity_comparison_demo\matrix_results.csv
```

생성되는 그림:

```text
experiments\reports\batch_runs\capacity_comparison_demo\figures\capacity_config_comparison.png
```

보고서에 사용한 기존 결과는 다음 경로에 보관되어 있다.

```text
experiments\reports\batch_runs\linter_comparison_default_20260618
experiments\reports\batch_runs\capacity_comparison_20260618
```

## 5. 시간이 부족할 때의 축약 시연

시뮬레이션 데이터가 이미 생성되어 있다면 다음 세 명령만 실행한다.

```powershell
cmd /c npx --yes squawk-cli@2.58.0 experiments\ddl\021_comparison_linter_pass_dynamic_danger.sql

.\migraguard.exe analyze experiments\ddl\021_comparison_linter_pass_dynamic_danger.sql --sandbox exp_06_heavy_log_insert.db --config experiments\configs\cases\default.yaml --output console

.\migraguard.exe analyze experiments\ddl\021_comparison_linter_pass_dynamic_danger.sql --sandbox exp_06_heavy_log_insert.db --config experiments\configs\cases\highcapacity.yaml --output console
```

## 6. 질문 대응

### 왜 실제 운영 DB에서 수집하지 않았는가?

> 실제 운영 DB에서 위험한 DDL을 반복 실행하여 정답 데이터를 만드는 것은 장애 위험이 있습니다. 따라서 이번 프로젝트에서는 모델의 동작 방향과 민감도를 재현 가능하게 검토하기 위해 시뮬레이션 데이터를 사용했습니다. 실제 운영 지표 수집과 결과 보정은 후속 검증 범위입니다.

### 동시 인덱스 생성 시간이 100ms인 것이 현실적인가?

> 현재 프로토타입은 테이블 재작성 여부를 중심으로 수행 시간을 추정하므로 동시 인덱스 생성의 테이블 스캔과 구축 비용을 충분히 반영하지 못합니다. 현재 값은 실제 수행 시간의 정확한 예측값이 아니라 모델 계산 결과이며, 작업 유형별 비용 모델로 보완해야 합니다.

### Default에서는 Danger인데 High Capacity에서는 Safe인 것이 맞는가?

> MigraGuard는 DDL에 고정 등급을 부여하는 도구가 아닙니다. 동일한 DDL이라도 현재 요청량, 활성 커넥션, 응답 지연과 서버 처리 여력에 따라 위험도가 달라지는 것이 모델의 의도입니다.

### 실제 장애를 예측할 수 있는가?

> 현재 실험은 실제 장애 예측 정확도를 증명한 것이 아닙니다. 정적 특성, 작업 부하와 서버 처리 여력을 결합했을 때 모델이 의도한 방향으로 반응하는지를 확인한 프로토타입 단계입니다.

## 7. 원격 접속 및 장애 대비

- 발표 전에 데스크톱 절전 모드와 자동 업데이트를 해제한다.
- 원격 접속은 발표 시작 전에 연결해 둔다.
- 노트북 테더링 등 예비 네트워크를 준비한다.
- 노트북에도 저장소, `migraguard.exe`, Go 실행 환경을 준비한다.
- Squawk는 최초 실행 시 패키지를 내려받을 수 있으므로 발표 전에 한 번 실행한다.
- `exp_06_heavy_log_insert.db`를 미리 생성해 둔다.
- Squawk 0 issues, Default Danger, High Capacity Safe 결과 화면을 캡처해 둔다.
- 실시간 실행이 실패하면 캡처 화면과 보고서의 그림 2, 그림 3, 그림 6을 이용해 흐름을 설명한다.

## 8. 발표 직전 체크리스트

- [ ] 데스크톱 원격 접속 확인
- [ ] 프로젝트 루트에서 터미널 실행
- [ ] `migraguard.exe` 실행 확인
- [ ] Squawk 명령 사전 실행
- [ ] `exp_06_heavy_log_insert.db` 생성 확인
- [ ] Default 결과가 Danger인지 확인
- [ ] High Capacity 결과가 Safe인지 확인
- [ ] 결과 화면 캡처 준비
- [ ] PDF 보고서의 그림 2, 그림 3, 그림 6 위치 확인
