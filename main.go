package main

import (
	"fmt"
	"os"

	"lem-in/internal/lem"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("ERROR: invalid data format")
		return
	}

	filePath := os.Args[1]
	res, err := lem.RunFile(filePath)
	if err != nil {
		fmt.Println("ERROR: invalid data format")
		return
	}

	fmt.Print(res)
}
