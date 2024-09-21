package main

func getChunkAtOffset(reader *mmap.ReaderAt, offset int) ([]byte, int) {
	bufferSize := 1 << 16
	buf := make([]byte, bufferSize)

	_, err := reader.ReadAt(buf, int64(offset))
	if err == io.EOF { return buf, -1 }

	// Identify the most recent full line, and set offset to that position +1 
	// to avoid newlines in the next chunk
	bufferOffset := 0
	for bufferOffset = bufferSize-1; bufferOffset>=0; bufferOffset-- {
		b := buf[bufferOffset]
		if b == '\n' {
			offset += bufferOffset+1
			break
		}
	}
	// Return the chunk of data and the position of the end of this chunk
	return buf[:bufferOffset], offset
}

func processChunk(chunk []byte, hm *HashMap) {
	var offset int
	var name []byte
	var temp []byte
	var f int
	var ok bool
	var tempData *TempData
	var i = 0
	for ; i<=len(chunk); i++ {
		if chunk[i] == '\n' {
			name, temp = split(chunk[offset:i])
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
			offset = i+1
		}
	}
}

func readLinesFromChunk(chunk []byte) [][]byte {
	var lines [][]byte
	var offset int
	i:=0
	for ; i<len(chunk); i++ {
		if chunk[i] == '\n' {
			lines = append(lines, chunk[offset:i])
			offset = i+1
		}
	}
	return lines
}

func readChunkedTempData(reader *mmap.ReaderAt, hm *HashMap) {
	offset:=0
	var chunk []byte
	lastVal := 0.0
	for {
		printProgress(offset, reader.Len(), &lastVal)
		chunk, offset = getChunkAtOffset(reader, offset)
		if offset == -1 { break }
		processChunk(chunk, hm)
	}
}

func printProgress(offset int, totalSize int, lastValue *float64) {
	completion := 100.0 * float64(offset) / float64(totalSize)
	if completion > *lastValue + .1 {
		fmt.Printf("%.1f%%\n", completion)
		*lastValue = completion
	}
}

func processTempData(hm *HashMap) {
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

