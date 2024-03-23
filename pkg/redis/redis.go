package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

const (
	redisHost = "redis.host"
	redisPort = "redis.port"
	redisUser = "redis.user"
	redisPass = "redis.password"
)

const servicePrefix = "awesome_service." // наш префикс сервиса

type RedisClient struct {
	config RedisConfig
	logger *logrus.Entry
	client *redis.Client
}

type Client interface {
	CheckJWTInBlacklist(ctx context.Context, jwtStr string) error
	WriteJWTToBlacklist(ctx context.Context, jwtStr string, jwtTTL time.Duration) error
}

func InitRedisConfig(vp *viper.Viper) RedisConfig {
	return RedisConfig{
		Host:        vp.GetString(redisHost),
		Password:    vp.GetString(redisPass),
		Port:        vp.GetInt(redisPort),
		User:        vp.GetString(redisUser),
		DialTimeout: time.Duration(vp.GetInt("redis.dialTimeout")) * time.Second,
		ReadTimeout: time.Duration(vp.GetInt("redis.readTimeout")) * time.Second,
	}
}

func NewRedisClient(ctx context.Context, config RedisConfig, logger *logrus.Logger) (*RedisClient, error) {
	client := &RedisClient{logger: logger.WithField("component", "redis")}

	client.config = config

	redisClient := redis.NewClient(&redis.Options{
		Password:    config.Password,
		Username:    config.User,
		Addr:        config.Host + ":" + strconv.Itoa(config.Port),
		DB:          0,
		DialTimeout: config.DialTimeout,
		ReadTimeout: config.ReadTimeout,
	})

	client.client = redisClient

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logger.Error("cant ping redis: ", err)
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return client, nil
}

func (c *RedisClient) Close() error {
	return c.client.Close()
}
