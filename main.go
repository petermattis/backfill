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
	username  string
)

var rootCmd = &cobra.Command{
	Use:   "backfill [command] (flags)",
	Short: "cockroach performance benchmarking backfill tool",
	Long: `backfill is a tool for building historical cockroach binaries and running
roachtests against them in order to backfill historical performance numbers.

The build command builds binaries and the run command runs tests. Historical
binaries are stored in --bin-dir (default ${HOME}/binaries) and test output is
stored in --artifacts (default ${HOME}/artifacts). Building the historical
binaries is time consuming. Using a GCE worker or other powerful machine is
recommended.

The backfill tool builds one historical binary per day using the last merge
commit on that day. The binaries are named <bin-dir>/cockroach-YYYYMMDD-<sha>.
Currently only binaries built from master are supported. The build command
requires that the current working directory be somewhere within the cockroach
repo and docker needs to be installed (the binaries are built as release
binaries using the builder docker image).

The run command runs the specified roachtests using the binaries it finds in
<bin-dir>. Test output data is stored in directories named
<artifacts>/YYYYMMDD. The test command requires that bin/roachtest and
bin/workload exist and that roachprod is somewhere in PATH.
`,
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

	runCmd.Flags().StringVarP(
		&username, "user", "u", username, "username to run under, detect if blank")

	if err := rootCmd.Execute(); err != nil {
		// Cobra has already printed the error message.
		os.Exit(1)
	}
}
