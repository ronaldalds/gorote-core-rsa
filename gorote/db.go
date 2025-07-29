package gorote

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type InitRedis struct {
	Host     string
	Password string
	Database int
	Port     int
}

type InitGorm struct {
	Host     string
	User     string
	Password string
	Database string
	Port     int
	TimeZone string
	Schema   string
}

type InitMongo struct {
	Host       string
	User       string
	Password   string
	Database   string
	Port       int
	AuthSource string
}

func (m *InitMongo) NewMongo() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/?authSource=%s&ssl=false",
		m.User, m.Password, m.Host, m.Port, m.AuthSource,
	)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	log.Println("Connected to MongoDB")
	return client, nil
}

func (r *InitRedis) NewRedis() (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := redis.NewClient(
		&redis.Options{Addr: fmt.Sprintf("%s:%d", r.Host, r.Port), Password: r.Password, DB: r.Database},
	)
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")
	return client, nil
}

func (g *InitGorm) NewGorm() (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s search_path=%s client_encoding=UTF8",
		g.Host, g.Port, g.User, g.Password, g.Database, g.TimeZone, g.Schema,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get *sql.DB from GORM: %v", err)
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	log.Println("Connected to Gorm and schema is set up.")
	return db, nil
}
