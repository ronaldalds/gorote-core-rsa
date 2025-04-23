package core

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
)

func (s *Service) HealthGorm() (map[string]string, error) {
	stats := make(map[string]string)
	// Cria um contexto com timeout para o health check
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if client := s.CoreGorm; client == nil {
		return nil, fmt.Errorf("failed to connect to Redis")
	}

	// Access the underlying *sql.DB from GORM and ping it
	sqlDB, err := s.CoreGorm.DB() // Obtém o *sql.DB subjacente do GORM
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db connection error: %v", err)
		log.Fatalf("db connection error: %v", err) // Log the error and terminate the program
		return stats, nil
	}

	err = sqlDB.PingContext(ctx) // Realiza o ping no banco de dados
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db ping failed: %v", err)
		log.Fatalf("db ping failed: %v", err) // Log the error and terminate the program
		return stats, nil
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Query for connection pool stats (PostgreSQL example)
	var dbStats struct {
		OpenConnections   int
		InUse             int
		Idle              int
		WaitCount         int64
		WaitDuration      time.Duration
		MaxIdleClosed     int64
		MaxLifetimeClosed int64
	}
	// You can write your own SQL query to fetch database stats
	sqlStats := `
		SELECT 
    		(SELECT count(*) FROM pg_stat_activity WHERE state = 'active') as open_connections,
    		(SELECT count(*) FROM pg_stat_activity WHERE state = 'idle') as idle,
    		(SELECT count(*) FROM pg_stat_activity WHERE wait_event IS NOT NULL) as wait_count
		`
	err = s.CoreGorm.Raw(sqlStats).Scan(&dbStats).Error
	if err != nil {
		log.Printf("Failed to retrieve db stats: %v", err)
	}

	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse) // You can calculate in_use based on your needs
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String() // You can get a duration in some databases
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	return stats, nil
}
