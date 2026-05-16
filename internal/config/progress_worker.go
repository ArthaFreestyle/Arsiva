package config

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/usecase"
)

func startProgressFlushWorker(u usecase.ProgressSessionUseCase, log *logrus.Logger) {
	ctx := context.Background()
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// Run once at startup, then every 15 minutes.
	flushExpired(ctx, u, log)
	for range ticker.C {
		flushExpired(ctx, u, log)
	}
}

func flushExpired(ctx context.Context, u usecase.ProgressSessionUseCase, log *logrus.Logger) {
	keys, err := u.ListExpiredSessionKeys(ctx)
	if err != nil {
		log.Errorf("flushExpired: list expired keys: %v", err)
		return
	}
	for _, key := range keys {
		if _, err := u.Finalize(ctx, key, "expired"); err != nil {
			// Log and continue — one bad session must not stop the batch.
			log.Errorf("flushExpired: finalize key=%s: %v", key, err)
		}
	}
}
