package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aniruddh/uptime-monitor/backend/config"
	"github.com/aniruddh/uptime-monitor/backend/internal/db"
	"github.com/aniruddh/uptime-monitor/backend/internal/db/sqlc"
	"github.com/aniruddh/uptime-monitor/backend/internal/handler"
	"github.com/aniruddh/uptime-monitor/backend/internal/repository/postgres"
	"github.com/aniruddh/uptime-monitor/backend/internal/routes"
	"github.com/aniruddh/uptime-monitor/backend/internal/scheduler"
	"github.com/aniruddh/uptime-monitor/backend/internal/service"
	"github.com/aniruddh/uptime-monitor/backend/pkg/httpclient"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)
	monitorRepo := postgres.NewMonitorRepository(queries)
	checkRepo := postgres.NewCheckRepository(queries)

	monitorService := service.NewMonitorService(monitorRepo, checkRepo)
	checkService := service.NewCheckService(checkRepo)

	pinger := httpclient.NewPinger(cfg.PingTimeout)
	sched := scheduler.New(monitorRepo, pinger, checkService.RecordCheck, cfg.SchedulerPoll)
	go sched.Run(ctx)

	router := newRouter(cfg, monitorService, checkService)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func newRouter(cfg config.Config, monitorService *service.MonitorService, checkService *service.CheckService) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.AllowedOrigin},
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: false,
	}))

	monitorHandler := handler.NewMonitorHandler(monitorService)
	checkHandler := handler.NewCheckHandler(checkService, monitorService)
	routes.Register(r, monitorHandler, checkHandler)

	return r
}
