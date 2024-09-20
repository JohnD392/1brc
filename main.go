package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime/pprof"
	"sort"
	"time"
)

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

func split(s []byte) ([]byte, []byte) {
	for i := 0; i < len(s); i++ {
		if s[i] == ';' {
			return s[:i], s[i+1:]
		}
	}
	return s, []byte{}
}

func readTempData(file *os.File, hm *HashMap) {
	count := 0
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		processLine(scanner.Bytes(), hm)
		count += 1
		if count%10000000 == 0 {
			fmt.Printf("read: %d lines, size: %d\n", count, hm.Size())
		}
	}
}

func processLine(line []byte, hm *HashMap) {
	name, temp := split(line)
	f := parseTemp(temp)
	tempData, ok := hm.Get(name)
	if !ok {
		nameCopy := make([]byte, len(name))
		copy(nameCopy, name)
		hm.Set(nameCopy, &TempData{
			name:  nameCopy,
			min:   f,
			max:   f,
			total: f,
			count: 1,
		})
	} else {
		if f < tempData.min {
			tempData.min = f
		}
		if f > tempData.max {
			tempData.max = f
		}
		tempData.total += f
		tempData.count++
	}
}

func process(hm *HashMap) {
	startTime := time.Now()
	var data []*TempData
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
	fmt.Printf("process time: %s\n", time.Since(startTime))
}

func attempt() {
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	hm := NewHashMap(1000)
	readTempData(file, hm)
	process(hm)
}
