package main

import (
	"context"
	"fmt"
	"log"

	"github.com/himclix/fuse/examples/service/internal/boot"
)

func main() {
	ctx := context.Background()

	if err := boot.Init(); err != nil {
		log.Fatalf("boot failed: %v", err)
	}

	cfg := boot.Cfg()

	fmt.Printf("\nStarting %s on :%s (env=%s)\n", cfg.App.ServiceName, cfg.App.Port, cfg.App.Env)
	fmt.Printf("DB: %s@%s:%d/%s (ssl=%s, pool=%d)\n",
		cfg.Database.Username, cfg.Database.Host, cfg.Database.Port,
		cfg.Database.Name, cfg.Database.SSLMode, cfg.Database.Pool.MaxOpenConnections)
	fmt.Printf("Kafka: brokers=%v tls=%v group=%s\n",
		cfg.Kafka.Brokers, cfg.Kafka.EnableTLS, cfg.Kafka.GroupID)

	_ = ctx
}
