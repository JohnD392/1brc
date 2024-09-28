package main

import (
	"bufio"
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
	fmt.Println(odds)
}

func InitializeMap(hm *HashMap) {
	file, _ := os.Open(filepath)
	scanner := bufio.NewScanner(file)
	cities := [][]byte{}
	for scanner.Scan() {
		b := scanner.Bytes()
		city, _ := split(b)
		cityName := make([]byte, len(city))
		copy(cityName, city)
		if !Contains(cities, cityName) {
			cities = append(cities, cityName)
		}
		if len(cities) == 413 {
			break
		}
	}
	if len(cities) != 413 {
		panic("Failed to find collisionless map")
	}
	for i := 0; i < len(cities); i++ {
		hm.Set(cities[i], &TempData{
			name:  cities[i],
			max:   10000,
			min:   -10000,
			count: 0,
			total: 0,
		})
	}
}

func search() int {
	file, _ := os.Open("cities.txt")
	scanner := bufio.NewScanner(file)
	cities := [][]byte{}
	for scanner.Scan() {
		city := scanner.Bytes()
		cities = append(cities, city)
	}
	for i := 12037; i < 50000; i++ {
		hashes := []int{}
		noCollisions := true
		for _, city := range cities {
			h := NewHashMap(i).hash(city)
			if ContainsInt(hashes, h) {
				noCollisions = false
				break
			} else {
				hashes = append(hashes, h)
			}
		}
		if noCollisions {
			sort.Ints(hashes)
			for _, h := range hashes {
				fmt.Println(h)
			}
			fmt.Println("Ideal size:", i)
			return i
		}
	}
	return 0
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
	tens := 0

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
	s := search()
	fmt.Println("S", s)
	hm := NewHashMap(s)
	InitializeMap(hm)
	readTempDataReader(file, hm)
	process(hm)
}
