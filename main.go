package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"sort"
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

func readTempDataReader(file *os.File, hm *HashMap) {
	var offset int64
	count := 0
	lastTime := time.Now()
	for {
		buffer := make([]byte, 1<<24)
		_, err := file.ReadAt(buffer, offset)
		lastNewlinePos := bytes.LastIndexByte(buffer, '\n')
		chunk := buffer[0:lastNewlinePos+1]
		offset += int64(lastNewlinePos + 1)
		for {
			newlineIndex := bytes.IndexByte(chunk, '\n')
			if newlineIndex == -1 {
				break
			}
			processLine(chunk[0:newlineIndex], hm)
			chunk = chunk[newlineIndex+1:]
			if len(chunk) <= 1 {
				break
			}
			count += 1
			if count%10000000 == 0 {
				pace := 100 * time.Since(lastTime) / 1000000000
				fmt.Printf("Pace: %d, hm size: %d\n", pace, hm.Size())
				lastTime = time.Now()
			}
		}
		if err == io.EOF {
			break
		}
		buffer = nil
	}
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
//	startTime := time.Now()
	var data []*TempData
	for _, v := range hm.buckets {
		if v != nil {
			data = append(data, v.Value)
		}
	}
	sort.Slice(data, func(a, b int) bool { return bytes.Compare(data[a].name, data[b].name) < 0 })
//	fmt.Print("{")
//	for i, v := range data {
//		if i > 0 {
//			fmt.Print(", ")
//		}
//		avgTemp := float64(v.total/v.count) / 10.0
//		fmt.Printf("%s=%.1f/%.1f/%.1f", v.name, float64(v.min)/10.0, avgTemp, float64(v.max)/10.0)
//	}
//	fmt.Print("}")
//	fmt.Printf("process time: %s\n", time.Since(startTime))
}

func attempt() {
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	hm := NewHashMap(10000)
	readTempDataReader(file, hm)
	process(hm)
}
