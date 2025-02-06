package maze

import (
	"fmt"
	"math/rand"
	"strings"
)

type CellPosition struct {
	Row int
	Col int
}

type Move struct {
	From      CellPosition
	To        CellPosition
	Direction string
}

type Maze struct {
	Width  int
	Height int
	Grid   [][]Cell
}

func New(width, height int) *Maze {
	grid := make([][]Cell, height)
	for i := range grid {
		grid[i] = make([]Cell, width)
		for j := range grid[i] {
			grid[i][j] = Cell{
				NorthWall: true,
				SouthWall: true,
				EastWall:  true,
				WestWall:  true,
				Reward:    0,
			}
		}
	}

	maze := &Maze{
		Width:  width,
		Height: height,
		Grid:   grid,
	}
	maze.generateMaze()
	return maze
}

func (m *Maze) randomCellPosition() CellPosition {
	return CellPosition{Row: rand.Intn(m.Height), Col: rand.Intn(m.Width)}
}

func (m *Maze) randomUnvisitedCellPosition(visited map[string]struct{}) CellPosition {
	for {
		pos := m.randomCellPosition()
		key := fmt.Sprintf("%d,%d", pos.Row, pos.Col)
		if _, included := visited[key]; !included {
			return pos
		}
	}
}

func (m *Maze) neighbors(pos CellPosition) []Move {
	directions := map[string]CellPosition{
		"North": {-1, 0}, "South": {1, 0}, "East": {0, 1}, "West": {0, -1},
	}
	var result []Move
	for dir, delta := range directions {
		neighbor := CellPosition{Row: pos.Row + delta.Row, Col: pos.Col + delta.Col}
		if neighbor.Row >= 0 && neighbor.Row < m.Height && neighbor.Col >= 0 && neighbor.Col < m.Width {
			result = append(result, Move{From: pos, To: neighbor, Direction: dir})
		}
	}
	return result
}

func (m *Maze) openWall(move Move) {
	// Open the wall in the given direction
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
}

func (m *Maze) randomWalk(visited map[string]struct{}) map[CellPosition]Move {
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

func (m *Maze) generateMaze() {

	visited := make(map[string]struct{})
	start := m.randomCellPosition()
	visited[fmt.Sprintf("%d,%d", start.Row, start.Col)] = struct{}{}

	for len(visited) < m.Width*m.Height {
		for cell, move := range m.randomWalk(visited) {
			m.openWall(move)
			visited[fmt.Sprintf("%d,%d", cell.Row, cell.Col)] = struct{}{}
		}
	}
}

func (m *Maze) String() string {
	var output string

	// Top boundary
	output += "+" + strings.Repeat("---+", m.Width) + "\n"

	for row := 0; row < m.Height; row++ {
		// Cell rows
		cellRow := "|"
		for col := 0; col < m.Width; col++ {
			cell := m.Grid[row][col]
			if !cell.EastWall {
				cellRow += "    "
			} else {
				cellRow += "   |"
			}
		}
		output += cellRow + "\n"

		// Wall rows
		wallRow := "+"
		for col := 0; col < m.Width; col++ {
			cell := m.Grid[row][col]
			if !cell.SouthWall {
				wallRow += "   +"
			} else {
				wallRow += "---+"
			}
		}
		output += wallRow + "\n"
	}

	return output
}

