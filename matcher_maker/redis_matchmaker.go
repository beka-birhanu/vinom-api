package matchmaker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/google/uuid"
)

const (
	// default prefix for redis key
	defaultPrefix string = "matchmaker"
	// default max player
	defaultMaxPlayer int64 = 2

	// defaultRankTolerance defines the maximum rank difference to match players
	defaultRankTolerance = 0

	// defaultLatencyTolerance defines the maximum latency difference to match players
	defaultLatencyTolerance = 0

	// list key string format
	queueRankLatencyKeyFmt string = "%s:queue:rank_%d:latency_%d"
)

// error types
var (
	ErrPlayerNotFoundInQueue = errors.New("player not found in queue")
)

// HandlerFunc is called when players matched
type HandlerFunc func(rank int, latency uint, IDs ...string)

type Player struct {
	ID      uuid.UUID
	Rank    int
	Latency uint
}

// Matchmaking options
type Options struct {
	// queue prefix
	Prefix string

	// Handler function to call when some players are matched
	Handler HandlerFunc

	// Matchmaker Logger
	Logger *log.Logger

	// MaxPlayer size for each match
	MaxPlayer int64

	// RankTolerance defines the maximum rank difference to match players
	RankTolerance int

	// LatencyTolerance defines the maximum latency difference to match players
	LatencyTolerance int
}

// Redis RedisMatchmaker manages the queue, pushing players to the queue and matching them with the same rank
type RedisMatchmaker struct {
	// Redis client
	client *redis.Client

	// Redis lock to lock the list when popping from it
	locker *redsync.Redsync

	// Matchmaker options
	opts *Options
}

// NewMatchmaker creates a new Matchmaker instance with provided Redis client and options
func NewMatchmaker(client *redis.Client, opts *Options) (*RedisMatchmaker, error) {
	if opts == nil {
		opts = &Options{
			MaxPlayer: defaultMaxPlayer,
			Prefix:    defaultPrefix,
		}
	}

	if opts.Logger == nil {
		opts.Logger = log.New(os.Stderr, fmt.Sprintf("%s: ", opts.Prefix), log.LstdFlags|log.Lshortfile)
	}

	if opts.MaxPlayer <= 0 {
		opts.MaxPlayer = defaultMaxPlayer
	}

	if opts.RankTolerance < 0 {
		opts.RankTolerance = defaultRankTolerance
	}

	if opts.LatencyTolerance < 0 {
		opts.LatencyTolerance = defaultLatencyTolerance
	}

	matchmaker := &RedisMatchmaker{
		opts:   opts,
		client: client,
	}
	pool := goredis.NewPool(matchmaker.client)
	matchmaker.locker = redsync.New(pool)
	return matchmaker, nil
}

// PushToQueue pushes a player to the queue
func (rm *RedisMatchmaker) PushToQueue(ctx context.Context, id uuid.UUID, rank int, latency uint) error {
	return rm.pushPlayerToQueue(ctx, &Player{
		id,
		rank,
		latency,
	})
}

func (rm *RedisMatchmaker) pushPlayerToQueue(ctx context.Context, player *Player) error {
	score := float64(time.Now().UnixNano())
	res := rm.client.ZAdd(ctx, rm.queueKey(player.Rank, player.Latency), &redis.Z{
		Score:  score,
		Member: player.ID,
	})
	_, err := res.Result()
	if err != nil {
		return err
	}
	go rm.match(ctx, player.Rank, player.Latency)
	return nil
}

// match locks the queue, queries the queue size, pops players if the size is greater than MaxPlayer, and calls the HandlerFunc
func (rm *RedisMatchmaker) match(ctx context.Context, rank int, latency uint) {
	queueKey := rm.queueKey(rank, latency)
	mutexKey := queueKey + ":match_lock"
	mutex := rm.locker.NewMutex(mutexKey)
	if err := mutex.Lock(); err != nil {
		rm.opts.Logger.Printf("error while obtaining match function lock: %s", err.Error())
	}

	defer func() {
		ok, err := mutex.Unlock()
		if err != nil {
			rm.opts.Logger.Printf("error while releasing match function lock: %s", err.Error())
		}

		if !ok {
			rm.opts.Logger.Printf("error while releasing match function lock: %s", "redis eval func returned 0 while releasing")
		}
	}()

	qLen := rm.client.ZCard(ctx, queueKey)
	if qLen.Val() >= rm.opts.MaxPlayer {
		result := rm.client.ZPopMin(ctx, queueKey, rm.opts.MaxPlayer)
		players := result.Val()

		var IDs []string
		for _, p := range players {
			IDs = append(IDs, fmt.Sprint(p.Member))
		}
		if rm.opts.Handler != nil {
			go rm.opts.Handler(rank, latency, IDs...)
		}
	}
}

func (rm *RedisMatchmaker) queueKey(rank int, latency uint) string {
	// Scalling rank and latency to increase group memeber
	return fmt.Sprintf(queueRankLatencyKeyFmt, rm.opts.Prefix, scale(rank, rm.opts.RankTolerance), scale(int(latency), rm.opts.LatencyTolerance))
}

// scale scales down the value based on the tolerance level
func scale(value, tolerance int) int {
	return value / (tolerance + 1)
}
