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

func calculateCollisionOdds(sampleSize, bucketSize int) {
	odds := float64(1)
	for i := 0; i < sampleSize; i++ {
		odds *= float64(bucketSize-i) / float64(bucketSize)
	}
	println(odds)
}

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
	tenths := int(s[len(s)-1]) - 48
	ones := int(s[len(s)-2]) - 48
	tens := 0b0

	isNegative := s[0] == '-'
	if isNegative {
		s = s[1:]
	}
	if len(s) > 3 {
		tens = int(s[0]) - 48
	}
	if isNegative {
		return -(tens*100 + ones*10 + tenths)
	}
	return tens*100 + ones*10 + tenths
}

func split(s []byte) ([]byte, []byte) {
	if s[len(s)-4] == ';' {
		return s[:len(s)-4], s[len(s)-3:]
	}
	if s[len(s)-5] == ';' {
		return s[:len(s)-5], s[len(s)-4:]
	}
	return s[:len(s)-6], s[len(s)-5:]
}

func readTempDataReader(file *os.File, hm *HashMap) {
	var offset int64
	count := 0
	lastTime := time.Now()
	for {
		buffer := make([]byte, 1<<24)
		_, err := file.ReadAt(buffer, offset)
		lastNewlinePos := bytes.LastIndexByte(buffer, '\n')
		chunk := buffer[0 : lastNewlinePos+1]
		offset += int64(lastNewlinePos + 1)
		for {
			newlineIndex := bytes.IndexByte(chunk[6:], '\n') + 6
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
				fmt.Printf("Pace: %d, count: %d, hm size: %d\n", pace, count, hm.Size())
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
}

func attempt() {
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	hm := NewHashMap(100000)
	readTempDataReader(file, hm)
	process(hm)
}
