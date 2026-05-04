# 📜 MigraGuard Scripts

이 디렉토리는 플랫폼 및 쉘 환경별 자동화 스크립트를 포함합니다.

### 📂 디렉토리 구조

*   **`cmd/`**: Windows 명령 프롬프트(CMD)용 `.bat` 스크립트
*   **`ps1/`**: Windows PowerShell용 `.ps1` 스크립트
*   **`sh/`**: Linux/macOS/Ubuntu(Bash)용 `.sh` 스크립트

### 🚀 주요 스크립트 설명

1.  **`sandbox.*`**: 시나리오 기반 샌드박스 생성 및 DDL 분석 통합 스크립트
2.  **`seed_all.*`**: 모든 시나리오에 대해 시뮬레이션 DB를 일괄 생성
3.  **`test_all.*`**: 전체 유닛 테스트 및 통합 테스트 실행

### 💡 사용 예시 (Windows CMD)
```cmd
:: 시뮬레이션 실행
scripts\cmd\sandbox.bat 03_spike.yaml 011_danger.sql

:: 전체 시드 생성
scripts\cmd\seed_all.bat
```
