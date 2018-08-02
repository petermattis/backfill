package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [from <YYYY-MM-DD>]",
	Short: "build binaries",
	Long:  ``,
	Args:  cobra.RangeArgs(0, 1),
	Run:   runBuild,
}

type target struct {
	date string
	sha  string
}

func (t target) String() string {
	return fmt.Sprintf("cockroach-%s-%s", t.date, t.sha)
}

func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func getTargets(from time.Time) []target {
	gitLog := `git log --pretty=format:"%ad%x09%H" --merges --since=` +
		from.Format("2006-01-02") + ` --date=iso-local master | ` +
		`awk '{print $NF, $1}' | uniq -f 1`
	cmd := exec.Command(`/bin/bash`, `-c`, gitLog)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	var results []target
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		if len(parts) != 2 {
			log.Fatalf("unexpected git log output: %s", scanner.Text())
		}
		results = append(results, target{
			date: strings.Replace(parts[1], "-", "", -1),
			sha:  parts[0],
		})
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return results
}

func run(args ...string) error {
	fmt.Printf("> %s\n", strings.Join(args, " "))
	if !dryRun {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}
	return nil
}

func mustRun(args ...string) {
	if err := run(args...); err != nil {
		log.Fatal(err)
	}
}

func gitTopLevel() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func buildOne(target target) {
	path := filepath.Join(binDir, target.String())
	if exists(path) {
		return
	}

	fmt.Printf("building %s\n", path)
	mustRun("git", "checkout", target.sha)
	mustRun("git", "clean", "-dxf")
	mustRun("git", "submodule", "update", "--init", "--force")

	const bin = "cockroach-linux-2.6.32-gnu-amd64"
	if err := os.Remove(bin); err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	if exists("build/builder/mkrelease.sh") {
		mustRun("build/builder.sh", "mkrelease", "amd64-linux-gnu")
	} else {
		mustRun("build/builder.sh", "make", "build", "TYPE=release-linux-gnu")
	}

	if exists(bin) {
		if err := os.Rename(bin, path); err != nil {
			log.Fatal(err)
		}
	}
}

func runBuild(_ *cobra.Command, args []string) {
	if err := os.Chdir(gitTopLevel()); err != nil {
		log.Fatal(err)
	}

	from := time.Date(2018, 4, 1, 0, 0, 0, 0, time.Local)
	if len(args) > 0 {
		var err error
		from, err = time.Parse("2006-01-02", args[0])
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := os.MkdirAll(binDir, 0755); err != nil {
		log.Fatal(err)
	}

	for _, target := range getTargets(from) {
		buildOne(target)
	}
}
