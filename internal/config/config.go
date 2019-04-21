package config

import (
	"context"
	"database/sql"
	"log"
	"time"

	"cloud.google.com/go/profiler"
	"cloud.google.com/go/storage"
	"github.com/acoshift/configfile"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

var config = configfile.NewReader("config")

func String(name string) string {
	return config.String(name)
}

var (
	redisClient   *redis.Client
	db            *sql.DB
	storageClient *storage.Client
)

func init() {
	var err error

	ctx := context.Background()

	redisClient = redis.NewClient(&redis.Options{
		Addr:        String("redis_addr"),
		Password:    String("redis_password"),
		IdleTimeout: 30 * time.Minute,
		MaxRetries:  3,
	})

	db, err = sql.Open("postgres", String("db_dsn"))
	must(err)

	storageClient, err = storage.NewClient(ctx)
	must(err)

	if config.Bool("profiling") {
		profiler.Start(profiler.Config{
			Service: "pikkanode",
		})
	}
}

func must(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func RedisClient() *redis.Client {
	return redisClient
}

func RedisPrefix() string {
	return String("redis_prefix")
}

func DB() *sql.DB {
	return db
}

func StorageClient() *storage.Client {
	return storageClient
}

func Dev() bool {
	return config.Bool("dev")
}

func BaseURL() string {
	return config.String("base_url")
}
