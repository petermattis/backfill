package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [tests]",
	Short: "run tests",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run:   runRun,
}

func runOne(i int, bin, tests string) {
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
	_ = os.RemoveAll(tmp)

	mustRun(`bin/roachtest`,
		`run`, tests,
		`--artifacts=`+tmp,
		`--cluster-id=`+fmt.Sprint(i),
		`--cockroach=`+bin,
		`--workload=bin/workload`,
		`--user=`+username)

	if !dryRun {
		if err := os.Rename(tmp, dest); err != nil {
			log.Fatal(err)
		}
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

	binaryPrefixFromTime := func(timeS string) string {
		t, err := time.Parse("2006-01-02", timeS)
		if err != nil {
			log.Fatal(err)
		}
		return "cockroach-" + t.Format("20060102") + "-"
	}

	if from != "" {
		from = binaryPrefixFromTime(from)
	}
	if to != "" {
		to = binaryPrefixFromTime(to)
	}

	ch := make(chan string, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			for {
				b, ok := <-ch
				if !ok {
					return
				}
				runOne(i+1, b, args[0])
			}
		}(i)
	}

	for _, b := range bins {
		if from != "" {
			base := filepath.Base(b)
			if base < from {
				continue
			}
			if to != "" && base > to {
				break
			}
		}

		ch <- b

		if count > 0 {
			count--
			if count == 0 {
				break
			}
		}
	}

	close(ch)
	wg.Wait()
}
