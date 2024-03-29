package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type wrappedObj struct {
	Type    uint16
	Payload interface{}
}

type RedisClient struct {
	*redis.Client
}

func (client *RedisClient) IsOk() bool {
	return client.Client != nil
}

func (client *RedisClient) InitRedis(ctx context.Context, addr string) error {
	// 创建Redis连接
	c := redis.NewClient(&redis.Options{
		Addr:     addr, // Redis服务器地址
		Password: "",   // Redis密码
		DB:       0,    // 使用默认的数据库
	})

	// 使用Ping命令检查连接是否正常
	pong, err := c.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("connecting to Redis err:%v", err)
	}
	log.Printf("Info: Connected to Redis:%v", pong)

	client.Client = c
	return nil
}

func (client *RedisClient) CloseRedis() {
	if !client.IsOk() {
		return
	}
	err := client.Close()
	if err != nil {
		log.Printf("Warning: closing connection err: %v\n", err)
	}
}

func (client *RedisClient) setRedis(ctx context.Context, key string, value string, expiration time.Duration) error {
	if !client.IsOk() {
		return fmt.Errorf("client is nil")
	}
	err := client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("setting key err: %v", err)
	}
	return nil
}

func (client *RedisClient) getRedis(ctx context.Context, key string) (string, int, error) {
	if client.Client == nil {
		return "", 0, fmt.Errorf("client is nil")
	}
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

func (client *RedisClient) StoreRedisCache(ctx context.Context, q Question, answers []RR) error {
	if !client.IsOk() {
		return fmt.Errorf("client is nil")
	}

	qStr, err := json.Marshal(q)
	if err != nil {
		return fmt.Errorf("marshaling q err: %v", err)
	}

	for _, a := range answers {
		wo := wrappedObj{
			Type:    a.Header().Rrtype,
			Payload: a,
		}
		value, err := json.Marshal(&wo)
		if err != nil {
			return fmt.Errorf("marshaling value err: %v", err)
		}

		// 当前时间戳
		now := time.Now().Unix()

		err = client.ZAdd(ctx, string(qStr), &redis.Z{
			Score:  float64(now + int64(a.Header().Ttl)),
			Member: string(value),
		}).Err()
		if err != nil {
			return fmt.Errorf("ZAdd err: %v", err)
		}
	}

	return nil
}

func (client *RedisClient) GetRedisCacheByKey(ctx context.Context, q Question) ([]RR, error) {
	if !client.IsOk() {
		return nil, fmt.Errorf("client is nil")
	}

	qStr, err := json.Marshal(q)
	if err != nil {
		return nil, fmt.Errorf("marshaling q err: %v", err)
	}

	// 当前时间戳
	now := time.Now().Unix()

	zList, err := client.ZRangeByScoreWithScores(ctx, string(qStr), &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", now),
		Max: "+inf",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("ZRangeByScoreWithScores err: %v", err)

	}

	var answers []RR
	for _, z := range zList {
		var wo wrappedObj
		zm, ok := z.Member.(string)
		if !ok {
			return nil, fmt.Errorf("unable to convert zm: %v", z.Member)
		}
		err = json.Unmarshal([]byte(zm), &wo)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling z.member err: %v", err)
		}

		if rrFunc, ok := TypeToRR[wo.Type]; ok {
			payloadStr, err := json.Marshal(wo.Payload)
			if err != nil {
				return nil, fmt.Errorf("marshaling payload err: %v", err)
			}
			rr := rrFunc()
			err = json.Unmarshal(payloadStr, &rr)
			if err != nil {
				return nil, fmt.Errorf("unmarshaling payload err: %v", err)
			}

			rr.Header().Ttl = uint32(z.Score - float64(now))

			answers = append(answers, rr)
		} else {
			return nil, fmt.Errorf("unsupported rr type %d", wo.Type)
		}
	}

	return answers, nil
}

func (client *RedisClient) GetRedisCacheAllData(ctx context.Context) (map[Question][]RR, error) {
	if !client.IsOk() {
		return nil, fmt.Errorf("client is nil")
	}

	keyList, err := client.Keys(ctx, "*").Result()
	if err != nil {
		return nil, fmt.Errorf("client.GetRedis err: %v", err)
	}

	res := make(map[Question][]RR)

	for _, key := range keyList {
		var q Question
		err = json.Unmarshal([]byte(key), &q)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling key err: %v", err)
		}

		rrs, err := client.GetRedisCacheByKey(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("GetRedisCacheByKey err: %v", err)
		}

		res[q] = rrs
	}

	return res, nil
}

func (client *RedisClient) CronRefreshData(ctx context.Context) {
	if !client.IsOk() {
		return
	}

	var err error

	for {
		var cursor uint64

		for {
			var keys []string
			keys, cursor, err = client.Scan(ctx, cursor, "*", 10).Result()
			if err != nil {
				log.Printf("CronRefreshData Scan err: %v", err)
			}

			// 当前时间戳
			now := time.Now().Unix()
			for _, key := range keys {
				_, err = client.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", now)).Result()
				if err != nil {
					log.Printf("CronRefreshData ZRemRangeByScore err: %v", err)
				}
			}

			if cursor == 0 {
				break
			}
		}

		time.Sleep(1 * time.Second)
	}
}
