package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	artifacts string
	binDir    string
	count     int
	dryRun    bool
)

var rootCmd = &cobra.Command{
	Use:   "backfill [command] (flags)",
	Short: "cockroach performance benchmarking backfill tool",
	Long:  ``,
}

func main() {
	cobra.EnableCommandSorting = false
	rootCmd.AddCommand(
		buildCmd,
		runCmd,
	)

	for _, cmd := range []*cobra.Command{buildCmd, runCmd} {
		cmd.Flags().StringVarP(
			&artifacts, "artifacts", "a", os.ExpandEnv("${HOME}/artifacts"),
			"directory to test artifacts")
		cmd.Flags().StringVarP(
			&binDir, "bin-dir", "b", os.ExpandEnv("${HOME}/binaries"),
			"directory to store binaries")
		cmd.Flags().IntVarP(
			&count, "count", "c", 0, "maximum number of test runs to perform")
		cmd.Flags().BoolVarP(
			&dryRun, "dry-run", "n", dryRun, "dry run (don't build binaries or run tests)")
	}

	if err := rootCmd.Execute(); err != nil {
		// Cobra has already printed the error message.
		os.Exit(1)
	}
}
