package workload

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

// Engine for generating requests by executing a desired workload
type Engine struct {
	Stats      []Stat
	OutFile    string
	TargetTput int64
	Duration   time.Duration
	Wld        *Workload
}

// NewEngine creates a new Engine instance for executing the workload
func NewEngine(outfile string, tput int64, duration string, w *Workload) (*Engine, error) {
	dur, err := time.ParseDuration(duration)
	if err != nil {
		return nil, err
	}
	isValid := w.IsValid()
	if !isValid {
		return nil, errors.New("Workload is not valid")
	}
	return &Engine{OutFile: outfile, TargetTput: tput, Duration: dur, Wld: w}, nil
}

func (e *Engine) init_prop_map() map[int]RequestFunction {
	sort.Slice(e.Wld.ApiInfos, func(i, j int) bool { return e.Wld.ApiInfos[i].Proportion > e.Wld.ApiInfos[j].Proportion })
	proportion_map := make(map[int]RequestFunction)
	var last_proportion_val int
	for _, api := range e.Wld.ApiInfos {
		var i int
		for i = 0; i < api.Proportion; i += 1 {
			proportion_map[last_proportion_val+i] = api.Fn
		}
		last_proportion_val += i
	}
	return proportion_map
}

// RunOpenLoop executes the request in an open loop fashion
func (e *Engine) RunOpenLoop(ctx context.Context) {
	prop_map := e.init_prop_map()
	log.Println("Target throughput", e.TargetTput)
	// Launch stat collector channel
	stat_channel := make(chan Stat, e.TargetTput)
	done := make(chan bool)
	go func() {
		count := 0
		for stat := range stat_channel {
			count += 1
			if count%1000 == 0 {
				log.Println("Processed", count, "requests")
			}
			e.Stats = append(e.Stats, stat)
		}
		close(done)
	}()

	// Launch the request maker goroutine that launches a request every tick_val
	tick_every := float64(1e9) / float64(e.TargetTput)
	tick_val := time.Duration(int64(1e9 / float64(e.TargetTput)))
	log.Println("Ticking after every", tick_val)
	stop := make(chan bool)

	var wg sync.WaitGroup
	go func() {
		src := rand.NewSource(0)
		g := distuv.Poisson{100, src}
		timer := time.NewTimer(0 * time.Second)
		next := time.Now()
		for {
			select {
			case <-stop:
				return
			case <-timer.C:
				n := rand.Intn(100)
				fn := prop_map[n]
				wg.Add(1)
				go func() {
					defer wg.Done()
					stat := fn(ctx)
					stat_channel <- stat
				}()
				next = next.Add(time.Duration(g.Rand()*tick_every/100) * time.Nanosecond)
				waitt := next.Sub(time.Now())
				timer.Reset(waitt)
			}
		}
	}()

	// Sleep for the desired duration while the requests are launched in the background
	time.Sleep(e.Duration)
	stop <- true
	// Wait for all the launched routines to finish
	wg.Wait()
	// Finish gathering stats
	close(stat_channel)
	<-done
	log.Println("Finished all requests")
}

// PrintStats prints the collected statistics and writes individual request information to the provided outfile in CSV format.
func (e *Engine) PrintStats() error {
	var num_errors int64
	var num_reqs int64
	var sum_durations int64
	stat_strings := []string{}
	for _, stat := range e.Stats {
		num_reqs += 1
		if stat.IsError {
			num_errors += 1
		}
		sum_durations += stat.Duration
		stat_strings = append(stat_strings, fmt.Sprintf("%d,%d,%t", stat.Start, stat.Duration, stat.IsError))
	}

	fmt.Println("Total Number of Requests:", num_reqs)
	fmt.Println("Successful Requests:", num_reqs-num_errors)
	fmt.Println("Error Responses:", num_errors)
	fmt.Println("Average Latency:", float64(sum_durations)/float64(num_reqs))
	// Write to file
	header := "Start,Duration,IsError\n"
	data := header + strings.Join(stat_strings, "\n")
	f, err := os.OpenFile(e.OutFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(data)
	if err != nil {
		return err
	}
	return nil
}
