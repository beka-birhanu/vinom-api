package main

import (
	"fmt"

	"github.com/beka-birhanu/vinom-api/maze"
)

func main() {
	maz := maze.New(10, 10)
	maze.PopulateReward(maze.RewardModel{RewardOne: 1, RewardTwo: 5, RewardTypeProb: 0.9}, maz)
	fmt.Println(maz)
}
