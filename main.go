package main

import (
	"fmt"

	"github.com/beka-birhanu/vinom-api/game/maze"
)

func main() {
	maz, _ := maze.New(10, 10)
	fmt.Println(maz)
	if maze.PopulateReward(maze.RewardModel{RewardOne: 1, RewardTwo: 5, RewardTypeProb: 0.9}, maz) != nil {
		return
	}
	fmt.Println(maz)
}
