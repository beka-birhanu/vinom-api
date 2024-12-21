package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/google/uuid"
)

const (
	// Default prefix for Redis keys
	defaultPrefix string = "matchmaker"
	// Default maximum number of players per match
	defaultMaxPlayer int64 = 2
	// Default rank tolerance (maximum rank difference to match players)
	defaultRankTolerance = 0
	// Default latency tolerance (maximum latency difference to match players)
	defaultLatencyTolerance = 0
	// Format string for queue Redis keys
	queueRankLatencyKeyFmt string = "%s:queue:rank_%d:latency_%d"
)

// Error definitions
var (
	ErrPlayerNotFoundInQueue = errors.New("player not found in queue")
)

// HandlerFunc is the function called when players are successfully matched
type HandlerFunc func(IDs ...string)

// Player represents a player in the matchmaking system
type Player struct {
	ID      uuid.UUID // Unique identifier of the player
	Rank    int       // Player's rank used for matchmaking
	Latency uint      // Player's latency used for matchmaking
}

// Options represents configuration options for the matchmaker
type Options struct {
	Prefix           string      // Prefix for Redis queue keys
	Handler          HandlerFunc // Function to call when players are matched
	Logger           *log.Logger // Logger for debugging and operational logs
	MaxPlayer        int64       // Maximum number of players per match
	RankTolerance    int         // Maximum rank difference for matching
	LatencyTolerance int         // Maximum latency difference for matching
}

// Matchmaker manages matchmaking using a sorted queue
type Matchmaker struct {
	sortedQueue i.SortedQueue // Sorted queue implementation
	opts        *Options      // Matchmaker configuration options
}

// NewMatchmaker initializes a new Matchmaker instance
func NewMatchmaker(sortedQueue i.SortedQueue, opts *Options) (*Matchmaker, error) {
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

// PushToQueue adds a player to the matchmaking queue
func (mm *Matchmaker) PushToQueue(ctx context.Context, id uuid.UUID, rank int, latency uint) error {
	return mm.pushPlayerToQueue(ctx, &Player{
		ID:      id,
		Rank:    rank,
		Latency: latency,
	})
}

// pushPlayerToQueue enqueues a player into the matchmaking queue
func (mm *Matchmaker) pushPlayerToQueue(ctx context.Context, player *Player) error {
	// Use the current time as the score for sorting players in the queue
	score := float64(time.Now().UnixNano())
	err := mm.sortedQueue.Enqueue(ctx, mm.queueKey(player.Rank, player.Latency), score, player.ID)
	if err != nil {
		return err
	}

	// Attempt to match players asynchronously
	go mm.match(ctx, player.Rank, player.Latency)
	return nil
}

// match checks if enough players are available in the queue for a match
// and calls the handler function if a match is found
func (mm *Matchmaker) match(ctx context.Context, rank int, latency uint) {
	queueKey := mm.queueKey(rank, latency)

	qLen := mm.sortedQueue.Count(ctx, queueKey)
	if qLen >= mm.opts.MaxPlayer {
		rawPlayers, err := mm.sortedQueue.DequeTops(ctx, queueKey, mm.opts.MaxPlayer)

		if err != nil {
			mm.opts.Logger.Printf("error while obtaining match function lock: %s", err.Error())
			return
		}
		var playersIDs []string

		// Convert the raw player data into a list of string IDs
		for _, raw := range rawPlayers {
			if id, ok := raw.(string); ok {
				playersIDs = append(playersIDs, id)
			} else {
				fmt.Println("Error: non-string value found in queue")
			}
		}

		if mm.opts.Handler != nil {
			go mm.opts.Handler(playersIDs...)
		}
	}
}

// queueKey generates the Redis queue key based on player rank and latency
func (mm *Matchmaker) queueKey(rank int, latency uint) string {
	// Scale rank and latency to based on tolerance
	return fmt.Sprintf(queueRankLatencyKeyFmt, mm.opts.Prefix, scale(rank, mm.opts.RankTolerance), scale(int(latency), mm.opts.LatencyTolerance))
}

// scale adjusts a value based on the specified tolerance level
func scale(value, tolerance int) int {
	return value / (tolerance + 1)
}
