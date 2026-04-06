package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Homeria/MigraGuard/internal/simulation"
)

func main() {
	dsn := flag.String("db", "postgres://user:pass@localhost:5432/target_db", "Target PostgreSQL DSN")
	conns := flag.Int("conns", 10, "Number of concurrent workers")
	profile := flag.String("profile", "steady", "Traffic profile (steady, flash-sale, read-heavy)")
	flag.Parse()

	gen, err := simulation.NewLoadGenerator(*dsn, *conns, simulation.TrafficProfile(*profile))
	if err != nil {
		log.Fatalf("❌ 부하 생성기 초기화 실패: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	gen.Run(ctx)
}
