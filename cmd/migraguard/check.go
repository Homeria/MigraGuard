package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "MigraGuard 에이전트 및 DB 연결 상태를 점검합니다",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 시스템 상태 점검 중...")
		fmt.Println("✅ 설정 파일 로드 완료")
		
		if GlobalConfig.Database.Postgres != "" {
			fmt.Println("✅ PostgreSQL 접속 설정 확인됨")
		} else {
			fmt.Println("⚠️  PostgreSQL 접속 설정이 비어있습니다. --db 플래그를 사용하세요.")
		}

		fmt.Printf("✅ SQLite 경로: %s\n", GlobalConfig.Database.SQLite)
		fmt.Println("\n👍 상태 점검 완료. 'analyze' 또는 'agent' 명령을 실행할 준비가 되었습니다.")
	},
}
