package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"sort"
	"sync"
	"time"
)

// const filepath = "../1brc/short_measurements.txt"
const filepath = "../1brc/measurements.txt"

func main() {
	f, err := os.Create("cpu_profile.prof")
	if err != nil {
		panic("What I can't make a file now?")
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	startTime := time.Now()
	attempt()
	fmt.Printf("total time: %s\n", time.Since(startTime))
}

type TempData struct {
	name  []byte
	min   int
	max   int
	total int
	count int
}

func parseTemp(s []byte) int {
	tenths := 0
	isNegative := s[0] == '-'
	i := 0
	if isNegative {
		i = 1
	}
	for ; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			continue
		}
		rawValue := int(c) - 48
		tenths *= 10
		tenths += rawValue
	}
	if isNegative {
		return -tenths
	}
	return tenths
}

// Assumes ; is in slice
func split(s []byte) ([]byte, []byte) {
	i := bytes.IndexByte(s, ';')
	return s[:i], s[i+1:]
}

func readTempDataReader(file *os.File) *[]*HashMap {
	var offset int64
	var hms []*HashMap
	var wg sync.WaitGroup
	for {
		buffer := make([]byte, 1<<24)
		_, err := file.ReadAt(buffer, offset)
		lastNewlinePos := bytes.LastIndexByte(buffer, '\n')
		chunk := buffer[0 : lastNewlinePos+1]
		offset += int64(lastNewlinePos + 1)
		println("offset:", offset)
		hm := NewHashMap(10000)
		wg.Add(1)
		go processChunk(chunk, hm, &hms, &wg)
		if err == io.EOF {
			println("Done allocating chunks")
			break
		}
		buffer = nil
	}
	println("Waiting for goroutines")
	wg.Wait()
	return &hms
}

func processChunk(chunk []byte, hm *HashMap, hms *[]*HashMap, wg *sync.WaitGroup) {
	for {
		newlineIndex := bytes.IndexByte(chunk, '\n')
		if newlineIndex == -1 { break }
		processLine(chunk[0:newlineIndex], hm)
		chunk = chunk[newlineIndex+1:]
		if len(chunk) <= 1 { break }
	}
	*hms = append(*hms, hm)
	wg.Done()
}

func processLine(line []byte, hm *HashMap) {
	name, temp := split(line)
	f := parseTemp(temp)
	tempData, ok := hm.Get(name)
	if !ok {
		nameCpy := make([]byte, len(name))
		copy(nameCpy, name)
		hm.Set(nameCpy, &TempData{
			name:  nameCpy,
			min:   f,
			max:   f,
			total: f,
			count: 1,
		})
	} else {
		if f < tempData.min { tempData.min = f }
		if f > tempData.max { tempData.max = f }
		tempData.total += f
		tempData.count++
	}
}

func process(hm *HashMap) {
	var data []*TempData
	println("size", hm.Size())
	for _, v := range hm.buckets {
		if v != nil {
			data = append(data, v.Value)
		}
	}
	sort.Slice(data, func(a, b int) bool { return bytes.Compare(data[a].name, data[b].name) < 0 })
	fmt.Print("{")
	for i, v := range data {
		if i > 0 {
			fmt.Print(", ")
		}
		avgTemp := float64(v.total/v.count) / 10.0
		fmt.Printf("%s=%.1f/%.1f/%.1f", v.name, float64(v.min)/10.0, avgTemp, float64(v.max)/10.0)
	}
	fmt.Print("}")
}

func mergeMaps(hms *[]*HashMap) *HashMap {
	mainMap := NewHashMap(10000)
	for _, hm := range *hms {
		for _, keyVal := range hm.buckets {
			if keyVal == nil { continue }
			hmTempData := keyVal.Value
			td, ok := mainMap.Get(hmTempData.name)
			if !ok {
				mainMap.Set(hmTempData.name, hmTempData)
			} else {
				if hmTempData.min < td.min { td.min = hmTempData.min }
				if hmTempData.max > td.max { td.max = hmTempData.max }
				td.total += hmTempData.total
				td.count += hmTempData.count
			}
		}
	}
	return mainMap
}

func attempt() {
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	hms := readTempDataReader(file)
	hm := mergeMaps(hms)
	process(hm)
}
