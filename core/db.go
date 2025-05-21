package core

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

func NewMongoStore(m *InitMongo) *MongoStore {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin&ssl=false",
		m.User,
		m.Password,
		m.Host,
		m.Port,
		m.Database,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("failed to connect to MongoDB: %w", err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("failed to ping MongoDB: %w", err)
	}
	fmt.Println("Connected to MongoDB")
	db := client.Database(m.Database)
	return &MongoStore{
		Client:   client,
		Database: db,
	}
}

func NewRedisStore(r *InitRedis) *RedisStore {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", r.Host, r.Port),
		Password: r.Password,
		DB:       r.Database,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	fmt.Println("Connected to Redis")
	return &RedisStore{Client: client}
}

func NewGormStore(g *InitGorm) *GormStore {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s search_path=%s client_encoding=UTF8",
		g.Host,
		g.Port,
		g.User,
		g.Password,
		g.Database,
		g.TimeZone,
		g.Schema,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to the database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get *sql.DB from GORM:", err)
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		log.Fatal("failed to ping database:", err)
	}

	fmt.Println("Connected to Gorm and schema is set up.")

	return &GormStore{DB: db}

}
