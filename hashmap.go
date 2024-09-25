package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// KeyValue represents a key-value pair in the hashmap
type KeyValue struct {
	Key   []byte
	Value *TempData
}

// HashMap represents our hashmap structure
type HashMap struct {
	buckets    []*KeyValue
	size       int
	bucketSize int
}

// NewHashMap creates a new HashMap with the given bucket size
func NewHashMap(bucketSize int) *HashMap {
	return &HashMap{
		buckets:    make([]*KeyValue, bucketSize),
		size:       0,
		bucketSize: bucketSize,
	}
}

// hash generates a hash for the given key
func (hm *HashMap) hash(key []byte) int {
	if len(key) < 4 {
		return int(binary.BigEndian.Uint16(key) % uint16(hm.bucketSize))
	}
	return int(binary.BigEndian.Uint32(key) % uint32(hm.bucketSize))
}

// Set adds a key-value pair to the hashmap
func (hm *HashMap) Set(key []byte, value *TempData) {
	index := hm.hash(key)
	if hm.buckets[index] == nil {
		hm.buckets[index] = &KeyValue{Key: key, Value: value}
		hm.size++
	} else if bytes.Equal(hm.buckets[index].Key, key) {
		hm.buckets[index].Value = value
	} else {
		// Handle collision: find the next empty slot
		for i := (index + 1) % hm.bucketSize; i != index; i = (i + 1) % hm.bucketSize {
			if hm.buckets[i] == nil {
				hm.buckets[i] = &KeyValue{Key: key, Value: value}
				hm.size++
				return
			} else if bytes.Equal(hm.buckets[i].Key, key) {
				hm.buckets[i].Value = value
				return
			}
		}
		// If we get here, the hashmap is full
		fmt.Println("HashMap is full, cannot add more elements")
	}
}

// Get retrieves the value associated with the given key
func (hm *HashMap) Get(key []byte) (*TempData, bool) {
	index := hm.hash(key)
	if hm.buckets[index] != nil && bytes.Equal(hm.buckets[index].Key, key) {
		return hm.buckets[index].Value, true
	}
	// Handle collision: search for the key
	for i := (index + 1) % hm.bucketSize; i != index; i = (i + 1) % hm.bucketSize {
		if hm.buckets[i] == nil {
			break
		}
		if bytes.Equal(hm.buckets[i].Key, key) {
			return hm.buckets[i].Value, true
		}
	}
	return nil, false
}

// Delete removes a key-value pair from the hashmap
func (hm *HashMap) Delete(key []byte) bool {
	index := hm.hash(key)
	if hm.buckets[index] != nil && bytes.Equal(hm.buckets[index].Key, key) {
		hm.buckets[index] = nil
		hm.size--
		return true
	}
	// Handle collision: search for the key
	for i := (index + 1) % hm.bucketSize; i != index; i = (i + 1) % hm.bucketSize {
		if hm.buckets[i] == nil {
			break
		}
		if bytes.Equal(hm.buckets[i].Key, key) {
			hm.buckets[i] = nil
			hm.size--
			return true
		}
	}
	return false
}

// Size returns the number of elements in the hashmap
func (hm *HashMap) Size() int {
	return hm.size
}
