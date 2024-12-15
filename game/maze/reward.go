package maze

import (
	"fmt"
	"math"
	"math/rand"
)

// RewardModel defines the reward configuration for a maze.
// RewardOne and RewardTwo represent two possible reward values
// that can be assigned to maze cells.
// RewardTypeProb determines the base probability of assigning RewardOne
// over RewardTwo, adjusted dynamically based on cell location.
type RewardModel struct {
	RewardOne      int     // Value of the first reward type
	RewardTwo      int     // Value of the second reward type
	RewardTypeProb float32 // Base probability of RewardOne (0.0 to 1.0)
}

// PopulateReward assigns rewards to maze cells based on the RewardModel.
// The probability of assigning RewardTwo decreases as cells are closer
// to the center of the maze.
func PopulateReward(r RewardModel, m *WillsonMaze) error {
	if r.RewardTypeProb > 1 || r.RewardTypeProb < 0 || min(r.RewardOne, r.RewardTwo) < 0 {
		return fmt.Errorf("Invalid RewardModel")
	}

	visited := map[string]struct{}{}

	stack := []CellPosition{{0, 0}}
	startPosKey := "00"
	visited[startPosKey] = struct{}{}

	for len(stack) > 0 {
		cell := pop(&stack)
		m.Grid[cell.Row][cell.Col].Reward = r.RewardOne

		if rand.Float32() > calcProb(r.RewardTypeProb, cell, m.Width, m.Height) {
			m.Grid[cell.Row][cell.Col].Reward = r.RewardTwo
		}

		for _, nbr := range m.neighbors(cell) {
			key := fmt.Sprintf("%d,%d", nbr.To.Row, nbr.To.Col)
			if _, seen := visited[key]; !seen {
				visited[key] = struct{}{}
				stack = append(stack, nbr.To)
			}
		}
	}

	return nil
}

// pop removes and returns the last element of a stack of CellPositions.
// The stack is represented as a slice of CellPosition.
func pop(s *[]CellPosition) CellPosition {
	lastIndex := len(*s) - 1
	popped := (*s)[lastIndex]
	*s = (*s)[:lastIndex] // Remove the last element
	return popped
}

// calcProb calculates the adjusted probability of assigning RewardTwo
// based on the cell's distance from the center of the maze.
// The function "punishes" cells that are further from the center,
// meaning that as the distance to the center increases, the probability
// of assigning RewardTwo decreases.
func calcProb(p float32, cell CellPosition, mazeWidth, mazeHeight int) float32 {
	midRow, midCol := mazeHeight/2, mazeWidth/2

	// Calculate the Manhattan distance
	distToMid := math.Abs(float64(cell.Row-midRow)) + math.Abs(float64(cell.Col-midCol))
	maxDist := float64(midRow + midCol)

	// Normalize the distance and invert it
	normalizedDist := 1.0 - distToMid/maxDist

	// Scale the probability using the base value `p`
	prob := p + (1-p)*float32(normalizedDist)/10

	return prob
}
