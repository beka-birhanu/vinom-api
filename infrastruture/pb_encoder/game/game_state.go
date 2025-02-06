package gamepb

import "github.com/beka-birhanu/vinom-api/service/i"

var _ i.Maze = &Maze{}
var _ i.GameState = &GameState{}

// Maze-related functions

// Height implements game.Maze.
func (x *Maze) Height() int {
	return len(x.Grid)
}

// Width implements game.Maze.
func (x *Maze) Width() int {
	panic("unimplemented")
}

// GetTotalReward implements game.Maze.
func (x *Maze) GetTotalReward() int32 {
	panic("unimplemented")
}

// NewValidMove implements game.Maze.
func (x *Maze) NewValidMove(i.CellPosition, string) (i.Move, error) {
	panic("unimplemented")
}

// InBound implements game.Maze.
func (x *Maze) InBound(row int, col int) bool {
	panic("unimplemented")
}

// IsValidMove implements game.Maze.
func (x *Maze) IsValidMove(move i.Move) bool {
	panic("unimplemented")
}

// Move implements game.Maze.
func (x *Maze) Move(move i.Move) (int32, error) {
	panic("unimplemented")
}

// RemoveReward implements game.Maze.
func (x *Maze) RemoveReward(pos i.CellPosition) error {
	panic("unimplemented")
}

// RetriveGrid implements game.Maze.
func (x *Maze) RetriveGrid() [][]i.Cell {
	maze := make([][]i.Cell, 0)
	for _, row := range x.Grid {
		new_row := make([]i.Cell, 0)
		for _, cell := range row.Cells {
			new_row = append(new_row, cell)
		}
		maze = append(maze, new_row)
	}
	return maze
}

// SetGrid implements game.Maze.
func (x *Maze) SetGrid(g [][]i.Cell) {
	maze := make([]*Maze_Row, 0)
	for _, row := range g {
		maze_row := &Maze_Row{
			Cells: make([]*Cell, 0),
		}
		for _, cell := range row {
			maze_row.Cells = append(maze_row.Cells, cellFromInterface(cell))
		}
		maze = append(maze, maze_row)
	}
	x.Grid = maze
}

// GameState-related functions

// RetriveMaze implements game.GameState.
func (x *GameState) RetriveMaze() i.Maze {
	return x.GetMaze()
}

// RetrivePlayers implements game.GameState.
func (x *GameState) RetrivePlayers() []i.Player {
	players := make([]i.Player, 0)
	for _, p := range x.GetPlayers() {
		players = append(players, p)
	}
	return players
}

// SetMaze implements game.GameState.
func (x *GameState) SetMaze(m i.Maze) {
	x.Maze = mazeFromInterface(m)
}

// SetPlayers implements game.GameState.
func (x *GameState) SetPlayers(p []i.Player) {
	players := make([]*Player, len(x.Players))
	for _, player := range p {
		players = append(players, playerFromInterface(player))
	}
	x.Players = players
}

// SetVersion implements game.GameState.
func (x *GameState) SetVersion(v int64) {
	x.Version = v
}

// Helper functions for converting interfaces

// mazeFromInterface converts a game.Maze interface to a *Maze structure.
func mazeFromInterface(m i.Maze) *Maze {
	maze := &Maze{}
	maze.SetGrid(m.RetriveGrid())

	return maze
}

func gameStateFromInterface(gs i.GameState) *GameState {
	gameState := &GameState{}

	gameState.SetVersion(gs.GetVersion())
	gameState.SetMaze(gs.RetriveMaze())
	gameState.SetPlayers(gs.RetrivePlayers())

	return gameState
}
