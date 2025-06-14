package main

import (
	"context"
	"fmt"
	"os"

	"github.com/m-mizutani/t/pkg/cli"
)

func main() {
	ctx := context.Background()

	if err := cli.Run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
