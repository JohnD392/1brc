package main

import (
    "bufio"
    "fmt"
    "os"
    "runtime/pprof"
    "sort"
    "time"
)

const filepath = "../1brc/measurements.txt"

type TempData struct {
    name  string
    min   int
    max   int
    total int
    count int
}

func main() {
    // Set up profiler
    f, err := os.Create("cpu_profile.prof")
    if err != nil {
        panic("What I can't make a file now?")
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // Run
    startTime := time.Now()
    attempt()
	println(time.Since(startTime))
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

func attempt() {
    file, err := os.Open(filepath)
    if err != nil { panic(err) }
    defer file.Close()

	m := make(map[string]*TempData, 1000)
    readTempData(file, &m)
    process(&m)
}

func split(s []byte) ([]byte, []byte) {
    for i := 0; i < len(s); i++ {
        if s[i] == ';' {
            return s[:i], s[i+1:]
        }
    }
    return s, []byte{}
}

func binarySearch(arr []*TempData, target string) (*TempData, bool) {
    left, right := 0, len(arr)-1
	mid := 0
    for left <= right {
        mid = left + (right-left)/2
        if arr[mid].name == target {
            return arr[mid], true
        } else if arr[mid].name < target {
            left = mid + 1
        } else {
            right = mid - 1
        }
    }
    return nil, false
}

func readTempData(file *os.File, m *map[string]*TempData) {
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
		tempData, ok = (*m)[string(name)]
        if !ok {
			(*m)[string(name)] = &TempData{
                name:  string(name),
                min:   f,
                max:   f,
                total: f,
                count: 1,
            }
        } else {
            if f < tempData.min { tempData.min = f }
            if f > tempData.max { tempData.max = f }
            tempData.total += f
            tempData.count++
        }
        if count%10000000 == 0 {
            println(count)
        }
        count += 1
    }
}

func process(m *map[string]*TempData) {
    startTime := time.Now()
	var tempDataSlice []TempData
	for _, v := range *m {
		tempDataSlice = append(tempDataSlice, *v)
	}
	sort.Slice(tempDataSlice, func(a, b int) bool { return tempDataSlice[a].name < tempDataSlice[b].name })
    for _, i := range tempDataSlice {
        avgTemp := float64(i.total/i.count) / 10.0
        fmt.Printf("%s=%.1f/%.1f/%.1f\n", i.name, float64(i.min)/10.0, avgTemp, float64(i.max)/10.0)
    }
    fmt.Printf("process time: %v\n", time.Since(startTime))
}

