package main

import (
	"bufio"
	"time"
	"os"
	"strconv"
	"strings"
	"runtime/pprof"
	"fmt"
	"sort"
)

const filepath = "../1brc/measurements.txt"

type TempData struct {
	name string
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
	m := make(map[string]TempData, 1000)
	readTempData(file, &m)
	process(&m)
	println("Duration: ", time.Since(startTime).Truncate(time.Second).String())
}

func readTempData(file *os.File, m *map[string]TempData) {
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

		tempData, ok = (*m)[name]
		if !ok {
			(*m)[name] = TempData{
				name: name,
				min: f,
				max: f,
				total: f,
				count: 1,
			}
		} else {
			if f < tempData.min { tempData.min = f	}
			if f > tempData.max { tempData.max = f }
			tempData.total += f
			tempData.count++
			(*m)[name] = tempData
		}

		if count % 10000000 == 0 {
			println(count)
		}
		count += 1
    }
}

func process(m *map[string]TempData) {
	var data []TempData
	for _, v := range *m {
		data = append(data, v)
	}
	sort.Slice(data, func(a, b int) bool { return data[a].name > data[b].name }) 
	for _, i := range data {
		avgTemp := i.total / float64(i.count)
		fmt.Printf("%s=%.1f/%.1f/%.1f\n", i.name, i.min, avgTemp, i.max)
	}
}

