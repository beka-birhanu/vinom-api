package game

import "github.com/beka-birhanu/vinom-api/maze"

type GameState struct {
	maze        *maze.Maze
	players     []Players
	rewardsLeft int
}

type Players struct{}
