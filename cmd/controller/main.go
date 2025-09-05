package main

import (
	"fmt"
	"os"

	"github.com/ItsThatDude/namespaced-route-validator/pkg/buildinfo"
	"github.com/ItsThatDude/namespaced-route-validator/pkg/controller"
)

var (
	// VERSION set from Makefile.
	VERSION = buildinfo.DefaultVersion
)

func main() {
	fmt.Fprintf(os.Stdout, "Controller %v starting up", VERSION)

	if err := controller.Main(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
