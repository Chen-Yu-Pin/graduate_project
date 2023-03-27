package models

import (
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis"
)

type Redis struct {
	client *redis.Client
}

func NewRedisConnect() *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis_user:6379",
		Password: "******", // no password set
		DB:       0,        // use default DB
	})
	if client == nil {
		log.Println("can not connect to redis!")
		return nil
	}
	log.Println("connect to redis!")
	r := &Redis{
		client: client,
	}
	return r
}

func (c *Redis) SetVerifyCode(mail, code string) error {
	log.Println("try into insert code")
	result, err := c.client.Set(mail, code, 30*time.Minute).Result()
	if err != nil {
		return err
	}
	log.Println(result)
	return nil
}

func (c *Redis) GetVerifyCode(mail string) (*string, error) {
	value, err := c.client.Get(mail).Result()
	if err != nil {
		return nil, err
	}
	return &value, err
}
func (c *Redis) JoinToTokenList(userID, token string) error {
	log.Println(userID, token)
	_, err := c.client.ZAdd(userID, redis.Z{
		Member: token,
		Score:  float64(time.Now().Local().Add(24 * 60 * time.Minute).Unix()),
	}).Result()
	if err != nil {
		return err
	}
	return nil
}
func (c *Redis) SearchToken(userID, token string) error {
	score, err := c.client.ZScore(userID, token).Result()
	if err != nil {
		return err
	}
	if score == 0 {
		return errors.New("token dosn't exist")
	}
	return nil
}

func (c *Redis) DeleteToken(userID, token string) error {
	_, err := c.client.ZRem(userID, token).Result()
	if err != nil {
		return err
	}
	return nil
}
