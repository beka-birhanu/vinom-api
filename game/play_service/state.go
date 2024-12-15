package game

import (
	"sync"

	"github.com/beka-birhanu/vinom-api/game/maze"
	"github.com/google/uuid"
)

type GameState struct {
	maze         *maze.WillsonMaze
	players      map[uuid.UUID]Players // map of players indexed with their ID.
	rewardsLeft  int                   // tottal rewardsLeft.
	sync.RWMutex                       // Lock for thread safty.
}

type Players struct {
	pos maze.CellPosition
}
