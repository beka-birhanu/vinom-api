package main

import (
	"fmt"

	"github.com/beka-birhanu/vinom-api/maze"
)

func main() {
	maze := maze.New(10, 10)
	fmt.Println(maze)
}

