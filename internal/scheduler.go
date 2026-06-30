package internal

import (
	"context"
	"fmt"
	"keyclub-api/google"
	"keyclub-api/sync"
	"log/slog"

	"github.com/go-co-op/gocron/v2"
	"github.com/jmoiron/sqlx"
)

func StartScheduler(ctx context.Context, googleConfig google.GoogleConfig, memberSync *sync.SyncState, eventSync *sync.SyncState, db *sqlx.DB) (gocron.Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		slog.Error("scheduler: failed to create scheduler", "error", err)
		return nil, fmt.Errorf("failed to create scheduler: %v", err)
	}

	if err := sync.SyncMembers(ctx, googleConfig, memberSync, db); err != nil {
		slog.Error("scheduler: failed to sync members", "error", err)
		return nil, fmt.Errorf("failed to sync members: %v", err)
	}
	if err := sync.SyncEvents(ctx, googleConfig, eventSync, db); err != nil {
		slog.Error("scheduler: failed to sync events", "error", err)
		return nil, fmt.Errorf("failed to sync events: %v", err)
	}

	scheduler.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(0, 0, 0),
			),
		),
		gocron.NewTask(
			func() {
				if err := sync.SyncMembers(ctx, googleConfig, memberSync, db); err != nil {
					slog.Error("scheduler: failed to sync members", "error", err)
				}
				if err := sync.SyncEvents(ctx, googleConfig, eventSync, db); err != nil {
					slog.Error("scheduler: failed to sync events", "error", err)
				}
			},
		),
	)
	return scheduler, nil
}
