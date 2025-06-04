package core

import (
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type InitRabbitMQ struct {
	User     string
	Password string
	Host     string
	Port     int
	Vh       string
}

type LogLoki struct {
	AppName string
	Url     string
	Port    int
}

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
	Host     string
	User     string
	Password string
	Database string
	Port     int
}

type MongoStore struct {
	Client   *mongo.Client
	Database *mongo.Database
}

type RedisStore struct {
	*redis.Client
}

type GormStore struct {
	*gorm.DB
}

type LogTelemetry struct {
	Timestamp    string              `json:"timestamp"`
	Method       string              `json:"method"`
	Path         string              `json:"path"`
	Headers      map[string][]string `json:"headers"`
	IP           string              `json:"ip"`
	RequestBody  map[string]any      `json:"request_body"`
	Status       int                 `json:"status"`
	Latency      int64               `json:"latency"`
	ResponseBody string              `json:"response_body"`
}
