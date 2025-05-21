package core

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (s *GormStore) HealthGorm() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		stats := make(map[string]string)
		contx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if client := s; client == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to connect to Redis")
		}

		sqlDB, err := s.DB.DB()
		if err != nil {
			stats["status"] = "down"
			stats["error"] = fmt.Sprintf("db connection error: %v", err)
			log.Fatalf("db connection error: %v", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(stats)
		}

		err = sqlDB.PingContext(contx)
		if err != nil {
			stats["status"] = "down"
			stats["error"] = fmt.Sprintf("db ping failed: %v", err)
			log.Fatalf("db ping failed: %v", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(stats)
		}

		stats["status"] = "up"
		stats["message"] = "It's healthy"

		var dbStats struct {
			OpenConnections   int
			InUse             int
			Idle              int
			WaitCount         int64
			WaitDuration      time.Duration
			MaxIdleClosed     int64
			MaxLifetimeClosed int64
		}
		sqlStats := `
		SELECT 
    		(SELECT count(*) FROM pg_stat_activity WHERE state = 'active') as open_connections,
    		(SELECT count(*) FROM pg_stat_activity WHERE state = 'idle') as idle,
    		(SELECT count(*) FROM pg_stat_activity WHERE wait_event IS NOT NULL) as wait_count
		`
		err = s.DB.Raw(sqlStats).Scan(&dbStats).Error
		if err != nil {
			log.Printf("Failed to retrieve db stats: %v", err)
		}

		stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
		stats["in_use"] = strconv.Itoa(dbStats.InUse)
		stats["idle"] = strconv.Itoa(dbStats.Idle)
		stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
		stats["wait_duration"] = dbStats.WaitDuration.String()
		stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
		stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

		if dbStats.OpenConnections > 40 {
			stats["message"] = "The database is experiencing heavy load."
		}

		if dbStats.WaitCount > 1000 {
			stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
		}

		return ctx.Status(fiber.StatusOK).JSON(stats)
	}
}
