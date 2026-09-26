package main

import (
	"fmt"

	"github.com/dylanrclee/gator/internal/config"
)

func main() {
	read_result, err := config.Read()
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	err = read_result.SetUser("dylan")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	read_result, err = config.Read()
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	fmt.Print(read_result)
}
