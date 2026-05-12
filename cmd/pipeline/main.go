package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common"
	logger "github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/pipeline"
)

func main() {

	logConfig := logger.New()
	plog := logConfig.For("pipeline")
	flag.Parse()

	if len(os.Args) < 2 {
		log.Fatal("Usage: pipeline <command> [args]")
	}

	pipelineApp, err := common.GetPipelineApp(plog)
	if err != nil {
		log.Fatal(err)
	}

	plog.Info("Pipeline run", "Args", os.Args[1])

	// Cancels on SIGTERM — Cloud Run sends this before killing the container
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// TODO: read worker count from env (e.g. os.Getenv("WORKER_COUNT"))
	workerCount := 10

	syncPipeline := pipeline.NewPipeline(
		workerCount,
		func(ctx context.Context, job pipeline.SyncAccountsJob) error {
			return pipelineApp.PortfolioService.SyncUserAccounts(ctx, job.UserID)
		},
		func(ctx context.Context) ([]pipeline.SyncAccountsJob, error) {
			users := pipelineApp.UserService.GetUsers()
			// map users to jobs — one job per user
			jobs := make([]pipeline.SyncAccountsJob, len(users))
			for i, user := range users {
				jobs[i] = pipeline.SyncAccountsJob{UserID: user.ID}
			}

			return jobs, nil
		},
	)

	refreshPipeline := pipeline.NewPipeline(
		workerCount,
		func(ctx context.Context, job pipeline.RefreshPortfolioJob) error {
			return pipelineApp.PortfolioService.RefreshUserAccounts(ctx, job.UserID, job.Simulate)
		},
		func(ctx context.Context) ([]pipeline.RefreshPortfolioJob, error) {
			// TODO: fetch all users, map to []RefreshPortfolioJob
			users := pipelineApp.UserService.GetUsers()
			jobs := make([]pipeline.RefreshPortfolioJob, len(users))
			for i, user := range users {
				jobs[i] = pipeline.RefreshPortfolioJob{UserID: user.ID}
			}

			return jobs, nil
		},
	)

	switch os.Args[1] {
	case "sync-all":
		if err := syncPipeline.Run(ctx); err != nil {
			plog.Error("sync-all", "error", err)
			os.Exit(1)
		}

	case "refresh-all":
		if err := refreshPipeline.Run(ctx); err != nil {
			plog.Error("refresh-all", "error", err)
			os.Exit(1)
		}

	case "sync-user":
		if len(os.Args) < 3 {
			log.Fatal("Usage: pipeline sync-user <uid>")
			plog.Error("sync-user Usage: pipeline sync-user <uid>", "error", err)
			os.Exit(1)

		}
		if err := syncPipeline.RunForOne(ctx, pipeline.SyncAccountsJob{UserID: os.Args[2]}); err != nil {
			log.Fatal(err)
		}

	case "refresh-user":
		if len(os.Args) < 3 {
			plog.Error("refresh-user Usage: pipeline refresh-user <uid>", "error", err)
			os.Exit(1)
		}

		simulate := len(os.Args) > 3 && os.Args[3] == "--simulate"

		if err := refreshPipeline.RunForOne(ctx, pipeline.RefreshPortfolioJob{UserID: os.Args[2], Simulate: simulate}); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("Unknown command: %s", os.Args[1])
	}
}
