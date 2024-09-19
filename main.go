package main

import (
	"bufio"
	"time"
	"os"
	"strconv"
	"strings"
	"runtime/pprof"
	"fmt"
	"github.com/igrmk/treemap/v2"
)

const filepath = "../1brc/measurements.txt"

type TempData struct {
	min float64
	max float64
	total float64
	count int
}

func main() {
	// Set up profiler
	f, err := os.Create("cpu_profile.prof")
	if err != nil { panic("What I can't make a file now?") }
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// Run
	naive()
}

func naive() {
	startTime := time.Now()
	file, err := os.Open(filepath)
	if err != nil { panic(err) }
	tr := treemap.New[string, TempData]()
	readTempData(file, tr)
	process(tr)
	println("Duration: ", time.Since(startTime).Truncate(time.Second).String())
}

func readTempData(file *os.File, tr *treemap.TreeMap[string, TempData]) {
	count := 0
    scanner := bufio.NewScanner(file)
	var name string
	var temp string
	var parts []string
	var f float64
	var tempData TempData
	var ok bool
    for scanner.Scan() {
		parts = strings.Split(scanner.Text(), ";")
		name = parts[0]
		temp = parts[1]
		f, _ = strconv.ParseFloat(temp, 64)

		tempData, ok = tr.Get(name)
		if !ok {
			tr.Set(name, TempData{
				min: f,
				max: f,
				total: f,
				count: 1,
			})
		} else {
			if f < tempData.min { tempData.min = f	}
			if f > tempData.max { tempData.max = f }
			tempData.total += f
			tempData.count++
			tr.Set(name, tempData)
		}

		if count % 10000000 == 0 {
			println(count)
		}
		count += 1
    }
}

func process(tr *treemap.TreeMap[string, TempData]) {
	for it := tr.Iterator(); it.Valid(); it.Next() {
		name := it.Key()
		tempData := it.Value()
		avgTemp := tempData.total / float64(tempData.count)
		fmt.Printf("%s=%.1f/%.1f/%.1f\n", name, tempData.min, avgTemp, tempData.max)
	}
}

