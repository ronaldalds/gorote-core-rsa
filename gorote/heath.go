package gorote

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type Health struct {
	Status            string        `json:"status,omitempty"`
	Error             string        `json:"error,omitempty"`
	Message           string        `json:"message,omitempty"`
	OpenConnections   int           `json:"open_connections,omitempty"`
	InUse             int           `json:"in_use,omitempty"`
	Idle              int           `json:"idle,omitempty"`
	WaitCount         int64         `json:"wait_count,omitempty"`
	WaitDuration      time.Duration `json:"wait_duration,omitempty"`
	MaxIdleClosed     int64         `json:"max_idle_closed,omitempty"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed,omitempty"`
}

func HealthGorm(s *gorm.DB) (*Health, error) {
	var stats Health
	contx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if s == nil {
		return nil, fmt.Errorf("failed to connect to DB")
	}
	sqlDB, err := s.DB()
	if err != nil {
		stats.Status = "down"
		stats.Error = fmt.Sprintf("db connection error: %v", err)
		return &stats, nil
	}
	err = sqlDB.PingContext(contx)
	if err != nil {
		stats.Status = "down"
		stats.Error = fmt.Sprintf("db ping failed: %v", err)
		return &stats, nil
	}
	sqlStats := `SELECT 
			(SELECT count(*) FROM pg_stat_activity WHERE state = 'active') as open_connections,
			(SELECT count(*) FROM pg_stat_activity WHERE state = 'idle') as idle,
			(SELECT count(*) FROM pg_stat_activity WHERE wait_event IS NOT NULL) as wait_count
		`
	err = s.Raw(sqlStats).Scan(&stats).Error
	if err != nil {
		log.Printf("Failed to retrieve db stats: %v", err)
	}
	stats.Status = "up"
	stats.Message = "It's healthy"
	if stats.OpenConnections > 40 {
		stats.Message = "The database is experiencing heavy load."
	}
	if stats.WaitCount > 1000 {
		stats.Message = "The database has a high number of wait events, indicating potential bottlenecks."
	}
	return &stats, nil
}
