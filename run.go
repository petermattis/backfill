package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [tests]",
	Short: "run tests",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run:   runRun,
}

func runOne(bin, tests string) {
	base := filepath.Base(bin)
	parts := strings.Split(base, "-")
	if len(parts) != 3 {
		return
	}
	date := parts[1]
	dest := filepath.Join(artifacts, date)
	if exists(dest) {
		return
	}

	tmp := dest + ".tmp"
	run(`bin/roachtest`,
		`run`, `-u`, `peter`, tests,
		`--artifacts=`+tmp,
		`--workload=bin/workload`,
		`--cockroach=`+bin,
		`--cluster-id=1`)

	if err := os.Rename(tmp, dest); err != nil {
		log.Fatal(err)
	}
}

func runRun(cmd *cobra.Command, args []string) {
	if err := os.MkdirAll(artifacts, 0755); err != nil {
		log.Fatal(err)
	}

	bins, err := filepath.Glob(filepath.Join(binDir, "cockroach-*"))
	if err != nil {
		log.Fatal(err)
	}
	for i, b := range bins {
		if count > 0 && i >= count {
			break
		}
		runOne(b, args[0])
	}
}
