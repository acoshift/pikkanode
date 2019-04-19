package config

import (
	"database/sql"
	"log"
	"time"

	"github.com/acoshift/configfile"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

var config = configfile.NewReader("config")

func String(name string) string {
	return config.String(name)
}

var (
	redisClient *redis.Client
	db          *sql.DB
)

func init() {
	var err error

	redisClient = redis.NewClient(&redis.Options{
		Addr:        String("redis_addr"),
		Password:    String("redis_password"),
		IdleTimeout: 30 * time.Minute,
		MaxRetries:  3,
	})

	db, err = sql.Open("postgres", String("db_dsn"))
	must(err)
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

func Dev() bool {
	return config.Bool("dev")
}
