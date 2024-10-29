package bootstrap

import (
	"context"
	"fmt"

	"github.com/stackvity/aidoc-server/config"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB connects to the PostgreSQL database using pgxpool
func ConnectDB(cfg config.Config, log *zap.Logger) (*pgxpool.Pool, error) { // Inject the logger

	dbpool, err := pgxpool.New(context.Background(), config.DBURL(cfg))
	if err != nil {
		log.Error("Unable to connect to database", zap.Error(err))       // Use injected logger
		return nil, fmt.Errorf("failed to connect to database: %w", err) // Wrap the error
	}

	if err = dbpool.Ping(context.Background()); err != nil {
		log.Error("Couldn't ping database", zap.Error(err))        // Use injected logger
		return nil, fmt.Errorf("failed to ping database: %w", err) // Wrap the error
	}

	log.Info("Successfully connected to database") // Use injected logger
	return dbpool, nil
}
