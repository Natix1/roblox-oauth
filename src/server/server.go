package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type runContext = int

const (
	RunContextTesting runContext = iota
	RunContextProduction
)

type environmentData struct {
	ClientId      string
	ClientSecret  string
	ClientURI     string
	OAuthScope    string
	CORSDomain    string
	RedisAddress  string
	RedisPassword string
	RedisDatabase int

	RunContext runContext
}

var Environment environmentData
var RedisClient *redis.Client
var HttpClient *http.Client
var Logger *slog.Logger

func AssertEnvironmentValue(key string) string {
	value, found := os.LookupEnv(key)
	if !found {
		panic(fmt.Sprintf("Key %s not found in environment variables!", key))
	}

	return value
}

func init() {
	contextString := flag.String("c", "", "Defines the environment this server is ran on. Either 'prod' or 'test'")
	flag.Parse()

	var currentRunContext int

	switch strings.ToLower(*contextString) {
	case "prod":
		currentRunContext = int(RunContextProduction)
	case "test":
		currentRunContext = int(RunContextTesting)
	default:
		panic("Invalid context specified. Use --help for usage.")
	}

	godotenv.Load(".env")

	database, err := strconv.Atoi(AssertEnvironmentValue("REDIS_DB"))
	if err != nil {
		log.Fatal("Invalid database specified; string->int conversion failed.")
	}

	Environment = environmentData{
		ClientId:      AssertEnvironmentValue("CLIENT_ID"),
		ClientSecret:  AssertEnvironmentValue("CLIENT_SECRET"),
		ClientURI:     AssertEnvironmentValue("REDIRECT_URI"),
		OAuthScope:    AssertEnvironmentValue("OAUTH_SCOPE"),
		CORSDomain:    AssertEnvironmentValue("CORS_DOMAIN"),
		RedisAddress:  AssertEnvironmentValue("REDIS_ADDRESS"),
		RedisPassword: AssertEnvironmentValue("REDIS_PASSWORD"),
		RedisDatabase: database,

		RunContext: currentRunContext,
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     Environment.RedisAddress,
		Password: Environment.RedisPassword,
		DB:       Environment.RedisDatabase,
		Protocol: 3,

		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	context := context.Background()
	if rdb.Ping(context).Err() != nil {
		panic("Failed to connect to redis")
	}

	RedisClient = rdb
	HttpClient = &http.Client{
		Timeout: time.Second * 10,
	}

	options := &slog.HandlerOptions{}

	if currentRunContext == RunContextTesting {
		options.Level = slog.LevelDebug
	} else {
		options.Level = slog.LevelWarn
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, options))
	Logger = logger
}
