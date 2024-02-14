package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"math"

	"time"
)

type AnswerList struct {
	answers []RR
	ttl     uint32
}

type wrappedObj struct {
	Type    uint16
	Payload interface{}
}

type RedisClient struct {
	*redis.Client
}

func (client *RedisClient) InitRedis(addr string) {
	// 创建Redis连接
	c := redis.NewClient(&redis.Options{
		Addr:     addr, // Redis服务器地址
		Password: "",   // Redis密码
		DB:       0,    // 使用默认的数据库
	})

	// 使用Ping命令检查连接是否正常
	ctx := context.Background()
	pong, err := c.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Error connecting to Redis:", err)
		return
	}
	fmt.Println("Connected to Redis:", pong)

	client.Client = c
}

func (client *RedisClient) CloseRedis() {
	err := client.Close()
	if err != nil {
		fmt.Println("Error closing connection:", err)
		return
	}
}

func (client *RedisClient) setRedis(ctx context.Context, key string, value string, expiration time.Duration) error {
	err := client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("setting key err: %v", err)
	}

	log.Printf("setting key(%v) success", key)
	return nil
}

func (client *RedisClient) getRedis(ctx context.Context, key string) (string, int, error) {
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return "", 0, fmt.Errorf("getting value err: %v", err)
	}

	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		return "", 0, fmt.Errorf("getting ttl err: %v", err)
	}

	return val, int(ttl), nil
}

func (client *RedisClient) StoreRedisCache(key Question, answers []RR) {
	var ttl uint32
	ttl = math.MaxUint32
	for _, a := range answers {
		if ttl > a.Header().Ttl {
			ttl = a.Header().Ttl
		}
	}

	keyStr, err := json.Marshal(key)
	if err != nil {
		log.Printf("marshaling key err: %v", err)
		return
	}

	var woList []wrappedObj
	for _, a := range answers {
		woList = append(woList, wrappedObj{
			Type:    a.Header().Rrtype,
			Payload: a,
		})
	}

	woListStr, err := json.Marshal(&woList)
	if err != nil {
		log.Printf("marshaling answers err: %v", err)
	}

	err = client.setRedis(context.Background(), string(keyStr), string(woListStr), time.Duration(ttl)*time.Second)
	if err != nil {
		log.Printf("client.SetRedis err: %v", err)
		return
	}
}

func (client *RedisClient) GetRedisCache(key Question) (AnswerList, bool) {
	keyStr, err := json.Marshal(key)
	if err != nil {
		log.Printf("marshaling key err: %v", err)
		return AnswerList{}, false
	}

	woListStr, ttl, err := client.getRedis(context.Background(), string(keyStr))
	if err != nil {
		log.Printf("client.GetRedis err: %v", err)
		return AnswerList{}, false
	}

	var woList []wrappedObj
	err = json.Unmarshal([]byte(woListStr), &woList)
	if err != nil {
		log.Printf("unmarshaling key err: %v", err)
		return AnswerList{}, false
	}

	var answers []RR
	for _, wo := range woList {
		if rrFunc, ok := TypeToRR[wo.Type]; ok {
			payloadStr, err := json.Marshal(wo.Payload)
			if err != nil {
				log.Printf("marshaling payload err: %v", err)
			}
			rr := rrFunc()
			err = json.Unmarshal(payloadStr, &rr)
			if err != nil {
				log.Printf("unmarshaling payload err: %v", err)
			}
			answers = append(answers, rr)
		} else {
			log.Printf("unsupported rr type %d", wo.Type)
			return AnswerList{}, false
		}
	}

	return AnswerList{
		answers: answers,
		ttl:     uint32(ttl),
	}, true
}
