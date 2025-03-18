package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func cpuIntensiveTask(goroutines int, numbers []int) int64 {
	var result int64
	totalNums := len(numbers)
	stride := totalNums / goroutines

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(g int) {
			start := g * stride
			end := start + stride
			if g == goroutines-1 {
				end = totalNums
			}

			var partSum int
			for _, v := range numbers[start:end] {
				partSum += v
			}

			atomic.AddInt64(&result, int64(partSum))
			wg.Done()
		}(g)
	}

	wg.Wait()

	return result
}

type Stats struct {
	Average    time.Duration
	Min        time.Duration
	Max        time.Duration
	Median     time.Duration
	Samples    []time.Duration
	NumSamples int
}

func calculateStats(samples []time.Duration) Stats {
	if len(samples) == 0 {
		return Stats{}
	}

	var total time.Duration
	min := samples[0]
	max := samples[0]

	for _, d := range samples {
		total += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	// Calculate median
	sorted := make([]time.Duration, len(samples))
	copy(sorted, samples)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	median := sorted[len(sorted)/2]
	if len(sorted)%2 == 0 {
		median = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	}

	return Stats{
		Average:    total / time.Duration(len(samples)),
		Min:        min,
		Max:        max,
		Median:     median,
		Samples:    sorted,
		NumSamples: len(samples),
	}
}

func generateIntSlice(n int) []int {
	nums := make([]int, n)
	for i := 0; i < n; i++ {
		nums[i] = i
	}
	return nums
}

func runTest(numGoroutines int, numbers []int) time.Duration {
	start := time.Now()

	cpuIntensiveTask(numGoroutines, numbers)

	return time.Since(start)
}

func handler(w http.ResponseWriter, _ *http.Request) {
	const numTests = 1000 // Number of test runs
	samples := make([]time.Duration, numTests)
	numbers := generateIntSlice(10000000)
	numGoroutines := getNumGoroutines()
	for i := 0; i < numTests; i++ {
		samples[i] = runTest(numGoroutines, numbers)
	}

	stats := calculateStats(samples)

	fmt.Fprintf(w, "Test Results (across %d runs):\n", numTests)
	fmt.Fprintf(w, "MachineCPUs=%d AllocatableCPUs=4 GOMAXPROCS: %d\n", runtime.NumCPU(), runtime.GOMAXPROCS(0))
	fmt.Fprintf(w, "Number of Goroutines: %d\n", numGoroutines)
	fmt.Fprintf(w, "Average: %v\n", stats.Average)
	fmt.Fprintf(w, "Median:  %v\n", stats.Median)
	fmt.Fprintf(w, "Min:     %v\n", stats.Min)
	fmt.Fprintf(w, "Max:     %v\n", stats.Max)
}

func setGOMAXPROCS() {
	goMaxProcsEnv := os.Getenv("GOMAXPROCS")
	if goMaxProcsEnv == "" {
		return
	}
	goMaxProcs, err := strconv.Atoi(goMaxProcsEnv)
	if err != nil {
		log.Printf("GOMAXPROCS environment is incorrect: %v", err)
		return
	}
	runtime.GOMAXPROCS(goMaxProcs)
}

func getNumGoroutines() int {
	const defaultNumGoroutines = 1
	numGoroutinesEnv := os.Getenv("NUM_GOROUTINES")
	if numGoroutinesEnv == "" {
		return defaultNumGoroutines
	}
	numGoroutines, err := strconv.Atoi(numGoroutinesEnv)
	if err != nil {
		log.Printf("NUM_GOROUTINES environment is incorrect: %v", err)
		return defaultNumGoroutines
	}

	return numGoroutines
}

func main() {
	setGOMAXPROCS()
	http.HandleFunc("/", handler)
	log.Printf("Starting server with logical CPUs=%d and GOMAXPROCS=%d", runtime.NumCPU(), runtime.GOMAXPROCS(0))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
