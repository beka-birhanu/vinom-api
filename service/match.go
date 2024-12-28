package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/beka-birhanu/vinom-api/config"
	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/google/uuid"
)

const (
	defaultPrefix           = "matchmaker"
	defaultMaxPlayer        = 2
	defaultRankTolerance    = 0
	defaultLatencyTolerance = 0
	queueRankLatencyKeyFmt  = "%s:queue:rank_%d:latency_%d"
)

var (
	ErrPlayerNotFoundInQueue = errors.New("player not found in queue")
)

type handlerFunc func(IDs []uuid.UUID)

type player struct {
	ID      uuid.UUID
	Rank    int
	Latency uint
}

type Options struct {
	Prefix           string
	Handler          handlerFunc
	Logger           *log.Logger
	MaxPlayer        int64
	RankTolerance    int
	LatencyTolerance int
}

type Matchmaker struct {
	sortedQueue i.SortedQueue
	opts        *Options
}

func NewMatchmaker(sortedQueue i.SortedQueue, opts *Options) (i.Matchmaker, error) {
	if opts == nil {
		opts = &Options{
			MaxPlayer: defaultMaxPlayer,
			Prefix:    defaultPrefix,
		}
	}

	if opts.Logger == nil {
		opts.Logger = log.New(io.Discard, "", 0)
	}

	if opts.MaxPlayer <= 0 {
		opts.MaxPlayer = defaultMaxPlayer
	}

	if opts.Prefix == "" {
		opts.Prefix = defaultPrefix
	}

	if opts.RankTolerance < 0 {
		opts.RankTolerance = defaultRankTolerance
	}

	if opts.LatencyTolerance < 0 {
		opts.LatencyTolerance = defaultLatencyTolerance
	}

	return &Matchmaker{
		opts:        opts,
		sortedQueue: sortedQueue,
	}, nil
}

func (mm *Matchmaker) PushToQueue(ctx context.Context, id uuid.UUID, rank int, latency uint) error {
	mm.opts.Logger.Printf("%s[INFO]%s Adding player to queue: ID=%s Rank=%d Latency=%d", config.LogInfoColor, config.LogColorReset, id, rank, latency)
	return mm.pushPlayerToQueue(ctx, &player{
		ID:      id,
		Rank:    rank,
		Latency: latency,
	})
}

func (mm *Matchmaker) pushPlayerToQueue(ctx context.Context, player *player) error {
	score := float64(time.Now().UnixNano())
	err := mm.sortedQueue.Enqueue(ctx, mm.queueKey(player.Rank, player.Latency), score, player.ID.String())
	if err != nil {
		mm.opts.Logger.Printf("%s[ERROR]%s Failed to enqueue player: %s", config.LogErrorColor, config.LogColorReset, err)
		return err
	}

	mm.opts.Logger.Printf("%s[INFO]%s Player enqueued successfully: ID=%s", config.LogInfoColor, config.LogColorReset, player.ID)
	go mm.match(ctx, player.Rank, player.Latency)
	return nil
}

func (mm *Matchmaker) match(ctx context.Context, rank int, latency uint) {
	queueKey := mm.queueKey(rank, latency)
	qLen := mm.sortedQueue.Count(ctx, queueKey)

	if qLen >= mm.opts.MaxPlayer {
		rawPlayers, err := mm.sortedQueue.DequeTops(ctx, queueKey, mm.opts.MaxPlayer)
		if err != nil {
			mm.opts.Logger.Printf("%s[ERROR]%s Error obtaining match lock: %s", config.LogErrorColor, config.LogColorReset, err)
			return
		}

		var playersIDs []uuid.UUID
		for _, raw := range rawPlayers {
			if id, err := uuid.Parse(raw); err == nil {
				playersIDs = append(playersIDs, id)
			} else {
				mm.opts.Logger.Printf("%s[ERROR]%s Non-UUID value in queue: %s", config.LogErrorColor, config.LogColorReset, raw)
			}
		}

		if mm.opts.Handler != nil {
			mm.opts.Logger.Printf("%s[INFO]%s Match found for players: %v", config.LogInfoColor, config.LogColorReset, playersIDs)
			go mm.opts.Handler(playersIDs)
		}
	}
}

func (mm *Matchmaker) SetMatchHandler(f func([]uuid.UUID)) {
	mm.opts.Handler = f
}

func (mm *Matchmaker) queueKey(rank int, latency uint) string {
	return fmt.Sprintf(queueRankLatencyKeyFmt, mm.opts.Prefix, scale(rank, mm.opts.RankTolerance), scale(int(latency), mm.opts.LatencyTolerance))
}

func scale(value, tolerance int) int {
	return value / (tolerance + 1)
}
