package matchmaker

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

// RedisSortedQueue manages a sorted queue in Redis.
type RedisSortedQueue struct {
	client *redis.Client
	locker *redsync.Redsync
}

// NewRedisSortedQueue initializes a RedisSortedQueue with the provided Redis client.
func NewRedisSortedQueue(client *redis.Client) (*RedisSortedQueue, error) {
	queue := &RedisSortedQueue{client: client}
	pool := goredis.NewPool(client)
	queue.locker = redsync.New(pool)
	return queue, nil
}

// Enqueue adds a member to the sorted queue with a given score.
func (rsq *RedisSortedQueue) Enqueue(ctx context.Context, queueKey string, score float64, member interface{}) error {
	_, err := rsq.client.ZAdd(ctx, queueKey, &redis.Z{Score: score, Member: member}).Result()
	return err
}

// DequeTop removes and retrieves up to `amount` members with the lowest scores.
func (rsq *RedisSortedQueue) DequeTop(ctx context.Context, queueKey string, amount int64) ([]interface{}, error) {
	mutex := rsq.locker.NewMutex(queueKey + ":match_lock")
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.Unlock()
	}()

	var members []interface{}
	if rsq.client.ZCard(ctx, queueKey).Val() >= amount {
		for _, p := range rsq.client.ZPopMin(ctx, queueKey, amount).Val() {
			members = append(members, p.Member)
		}
	}
	return members, nil
}

// Count returns the number of members in the sorted queue.
func (rsq *RedisSortedQueue) Count(ctx context.Context, queueKey string) int64 {
	return rsq.client.ZCard(ctx, queueKey).Val()
}
