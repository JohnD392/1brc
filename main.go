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

	"golang.org/x/exp/mmap"
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
	//	attempt()
	//	offset := 0
	//	var buf []byte
	//	for {
	//		buf, offset = readDataFromFile(offset)
	//		if offset == -1 {
	//			break
	//		}
	//		println("Offset:", offset)
	//		println(string(buf[len(buf)-50:]))
	//		println("---------------------------------------------------------------------------")
	//	}
	file, _ := os.Open(filepath)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_ = scanner.Bytes()
	}
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

func readDataFromFile(offset int) ([]byte, int) {
	reader, err := mmap.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	bufferSize := 1 << 24 //16M

	buf := make([]byte, bufferSize)
	_, err = reader.ReadAt(buf, int64(offset))
	if err == io.EOF {
		// if we hit the end of the file, return the remaining bytes, and signal completion with offset=-1
		return buf, -1
	}

	bufferOffset := 0
	for bufferOffset = bufferSize - 1; bufferOffset >= 0; bufferOffset-- {
		b := buf[bufferOffset]
		if b == '\n' {
			offset += bufferOffset
			break
		}
	}
	return buf[:bufferOffset], offset
}

func readTempData(file *os.File, hm *HashMap) {
	count := 0
	scanner := bufio.NewScanner(file)
	var tempData *TempData
	var name []byte
	var temp []byte
	var f int
	var ok bool
	for scanner.Scan() {
		name, temp = split(scanner.Bytes())
		f = parseTemp(temp)

		tempData, ok = hm.Get(name)
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
		if count%10000000 == 0 {
			fmt.Printf("read: %d lines, size: %d\n", count, hm.Size())
		}
		count += 1
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
	for _, i := range data {
		avgTemp := float64(i.total/i.count) / 10.0
		fmt.Printf("%s=%.1f/%.1f/%.1f\n", i.name, float64(i.min)/10.0, avgTemp, float64(i.max)/10.0)
	}
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
