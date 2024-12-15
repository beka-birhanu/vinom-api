/*
Package maze provides tools for creating and managing rectangular mazes.

It defines the `Maze` structure, composed of `Cell` objects that include wall configurations
and optional rewards.

The package includes functionality for random maze generation with Wilson's algorithm, wall manipulation,
and reward assignment. Rewards can be dynamically distributed based on proximity to the maze center.

Utility functions enable neighbor detection, move validation, and ASCII visualization of the maze.
*/
package maze

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/beka-birhanu/vinom-api/game"
)

const (
	maxMazeDimenssion = 20
)

var (
	Directions = map[string]CellPosition{
		"North": {Row: -1, Col: 0},
		"South": {Row: 1, Col: 0},
		"East":  {Row: 0, Col: 1},
		"West":  {Row: 0, Col: -1},
	}

	ErrInvalidMove = errors.New("invalid move request")
)

// WillsonMaze represents a rectangular maze consisting of cells with walls and optional rewards.
type WillsonMaze struct {
	Width  int       // Width of the maze (number of columns)
	Height int       // Height of the maze (number of rows)
	Grid   [][]*Cell // 2D grid of cells forming the maze
}

// New initializes a new maze of the given dimensions and generates its layout.
func New(width, height int) (*WillsonMaze, error) {
	if min(width, height) <= 0 || max(width, height) > maxMazeDimenssion {
		return nil, fmt.Errorf("Invalid maze dimensions")
	}

	grid := make([][]*Cell, height)
	for i := range grid {
		grid[i] = make([]*Cell, width)
		for j := range grid[i] {
			grid[i][j] = &Cell{
				NorthWall: true,
				SouthWall: true,
				EastWall:  true,
				WestWall:  true,
				Reward:    0,
			}
		}
	}

	maze := &WillsonMaze{
		Width:  width,
		Height: height,
		Grid:   grid,
	}
	maze.generateMaze()
	return maze, nil
}

// randomCellPosition generates a random position within the maze.
func (m *WillsonMaze) randomCellPosition() CellPosition {
	return CellPosition{Row: rand.Intn(m.Height), Col: rand.Intn(m.Width)}
}

// randomUnvisitedCellPosition selects a random position that has not been visited.
func (m *WillsonMaze) randomUnvisitedCellPosition(visited map[string]struct{}) CellPosition {
	for {
		pos := m.randomCellPosition()
		key := fmt.Sprintf("%d,%d", pos.Row, pos.Col)
		if _, included := visited[key]; !included {
			return pos
		}
	}
}

// neighbors finds all valid moves from a given cell position.
func (m *WillsonMaze) neighbors(pos CellPosition) []Move {
	var result []Move
	for dir, delta := range Directions {
		neighbor := CellPosition{Row: pos.Row + delta.Row, Col: pos.Col + delta.Col}
		if neighbor.Row >= 0 && neighbor.Row < m.Height && neighbor.Col >= 0 && neighbor.Col < m.Width {
			result = append(result, Move{From: pos, To: neighbor, Direction: dir})
		}
	}
	return result
}

// openWall removes the wall between two adjacent cells in the specified direction.
func (m *WillsonMaze) openWall(move Move) error {
	switch move.Direction {
	case "North":
		m.Grid[move.From.Row][move.From.Col].NorthWall = false
		m.Grid[move.To.Row][move.To.Col].SouthWall = false
	case "South":
		m.Grid[move.From.Row][move.From.Col].SouthWall = false
		m.Grid[move.To.Row][move.To.Col].NorthWall = false
	case "East":
		m.Grid[move.From.Row][move.From.Col].EastWall = false
		m.Grid[move.To.Row][move.To.Col].WestWall = false
	case "West":
		m.Grid[move.From.Row][move.From.Col].WestWall = false
		m.Grid[move.To.Row][move.To.Col].EastWall = false
	}

	return nil
}

// randomWalk performs a random walk starting from an unvisited cell.
func (m *WillsonMaze) randomWalk(visited map[string]struct{}) map[CellPosition]Move {
	start := m.randomUnvisitedCellPosition(visited)
	visits := make(map[CellPosition]Move)
	cell := start

	for {
		neighbors := m.neighbors(cell)
		randomNeighbor := neighbors[rand.Intn(len(neighbors))]
		visits[cell] = randomNeighbor
		key := fmt.Sprintf("%d,%d", randomNeighbor.To.Row, randomNeighbor.To.Col)
		if _, included := visited[key]; included {
			break
		}
		cell = randomNeighbor.To
	}

	return visits
}

// generateMaze creates a maze using a randomized algorithm.
func (m *WillsonMaze) generateMaze() {
	visited := make(map[string]struct{})
	start := m.randomCellPosition()
	visited[fmt.Sprintf("%d,%d", start.Row, start.Col)] = struct{}{}

	for len(visited) < m.Width*m.Height {
		for cell, move := range m.randomWalk(visited) {
			_ = m.openWall(move)
			visited[fmt.Sprintf("%d,%d", cell.Row, cell.Col)] = struct{}{}
		}
	}
}

// IsValidMove checks if a move is valid (i.e., the connecting wall is down).
func (m *WillsonMaze) IsValidMove(move game.Move) bool {
	// Check if the move positions are within the maze bounds.
	inBound := func(row, col int) bool {
		return row >= 0 && row < m.Height && col >= 0 && col < m.Width
	}

	// Ensure both the starting and destination positions are valid.
	if !inBound(move.GetFrom().GetRow(), move.GetFrom().GetCol()) || !inBound(move.GetTo().GetRow(), move.GetTo().GetCol()) {
		return false
	}

	// Check the walls based on the direction of the move.
	switch move.GetDirection() {
	case "North":
		return !m.Grid[move.GetFrom().GetRow()][move.GetFrom().GetCol()].NorthWall && !m.Grid[move.GetTo().GetRow()][move.GetTo().GetCol()].SouthWall
	case "South":
		return !m.Grid[move.GetFrom().GetRow()][move.GetFrom().GetCol()].SouthWall && !m.Grid[move.GetTo().GetRow()][move.GetTo().GetCol()].NorthWall
	case "East":
		return !m.Grid[move.GetFrom().GetRow()][move.GetFrom().GetCol()].EastWall && !m.Grid[move.GetTo().GetRow()][move.GetTo().GetCol()].WestWall
	case "West":
		return !m.Grid[move.GetFrom().GetRow()][move.GetFrom().GetCol()].WestWall && !m.Grid[move.GetTo().GetRow()][move.GetTo().GetCol()].EastWall
	default:
		return false
	}
}

// Move executes a move in the maze, returning the reward of the destination cell.
func (m *WillsonMaze) Move(move game.Move) (int, error) {
	// Check if the move is valid.
	if !m.IsValidMove(move) {
		return 0, ErrInvalidMove
	}

	// Get the reward of the destination cell and reset the reward.
	reward := m.Grid[move.GetTo().GetRow()][move.GetTo().GetCol()].Reward
	m.Grid[move.GetTo().GetRow()][move.GetTo().GetCol()].Reward = 0

	return reward, nil
}

// String provides a textual representation of the maze.
func (m *WillsonMaze) String() string {
	var output string

	// Top boundary
	output += "+" + strings.Repeat("---+", m.Width) + "\n"

	for row := 0; row < m.Height; row++ {
		// Cell rows
		cellRow := "|"
		for col := 0; col < m.Width; col++ {
			cell := m.Grid[row][col]

			// Display reward if present, otherwise leave the cell empty
			if cell.Reward != 0 {
				cellRow += " " + fmt.Sprint(cell.Reward) + " "
			} else {
				cellRow += "   "
			}

			// Add east wall or space
			if cell.EastWall {
				cellRow += "|"
			} else {
				cellRow += " "
			}
		}
		output += cellRow + "\n"

		// Wall rows
		wallRow := "+"
		for col := 0; col < m.Width; col++ {
			cell := m.Grid[row][col]

			// Add south wall or space
			if cell.SouthWall {
				wallRow += "---+"
			} else {
				wallRow += "   +"
			}
		}
		output += wallRow + "\n"
	}

	return output
}
