package main

import (
	"fmt"
	"os"

	"github.com/ItsThatDude/namespaced-route-validator/pkg/controller"
)

func main() {
	if err := controller.Main(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
