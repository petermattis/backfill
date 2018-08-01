package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

var statsPathRE = regexp.MustCompile(`artifacts/(.*/[0-9]+\.logs/stats\.json)$`)

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

	tmp, err := ioutil.TempDir(artifacts, date+".tmp.")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		os.RemoveAll(tmp)
	}()

	run(`roachtest`,
		`run`, `-u`, `peter`, tests,
		`--artifacts=`+tmp,
		`--workload=bin/workload`,
		`--cockroach=`+bin,
		`--cluster-id=1`)

	err = filepath.Walk(tmp, func(path string, info os.FileInfo, err error) error {
		if !info.Mode().IsRegular() {
			return nil
		}
		m := statsPathRE.FindStringSubmatch(path)
		if m == nil {
			return nil
		}
		destPath := filepath.Join(dest, m[1])
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			log.Fatal(err)
		}
		if err := os.Rename(path, destPath); err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
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
	for _, b := range bins {
		runOne(b, args[0])
	}
}
