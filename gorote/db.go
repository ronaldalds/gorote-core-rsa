package gorote

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type InitRedis struct {
	User     string
	Password string
	Host     string
	Port     int
	Database int
}

func (d *InitRedis) URL() string {
	return fmt.Sprintf("redis://%s:%s@%s:%d/%d",
		d.User, d.Password, d.Host, d.Port, d.Database,
	)
}

type InitPostgres struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	TimeZone string
	Schema   string
}

func (d *InitPostgres) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s search_path=%s client_encoding=UTF8",
		d.Host, d.Port, d.User, d.Password, d.Database, d.TimeZone, d.Schema,
	)
}

type InitMySQL struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	TimeZone string
}

func (d *InitMySQL) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		d.User, d.Password, d.Host, d.Port, d.Database, d.TimeZone,
	)
}

type InitMongo struct {
	Host       string
	User       string
	Password   string
	Database   string
	Port       int
	AuthSource string
}

func (m *InitMongo) URI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=%s&ssl=false",
		m.User, m.Password, m.Host, m.Port, m.AuthSource,
	)
}

func NewMongo(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

func NewRedis(url string) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("error on parse url: %v", err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %v", err)
	}

	if err := client.Set(ctx, "foo", "bar", 0).Err(); err != nil {
		return nil, fmt.Errorf("error on set: %v", err)
	}

	val, err := client.Get(ctx, "foo").Result()
	if err != nil {
		return nil, fmt.Errorf("error on get: %v", err)
	}

	if val != "bar" {
		return nil, fmt.Errorf("error in operation")
	}
	log.Println("Connected to Redis")
	return client, nil
}

func NewGorm(dialector gorm.Dialector) (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get *sql.DB from GORM: %v", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Connected to Gorm and schema is set up.")
	return db, nil
}
