package main

import (
	"bytes"
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
	sum := 0
	for i := 0; i < len(key); i++ {
		sum += int(key[i])
		sum = sum << 5
	}
	if sum < 0 {
		return -sum % hm.bucketSize
	}
	return sum % hm.bucketSize
}

// Set adds a key-value pair to the hashmap
func (hm *HashMap) Set(key []byte, value *TempData) {
	index := hm.hash(key)
	hm.buckets[index] = &KeyValue{Key: key, Value: value}
	hm.size++
}

var CollisionCities []string

func ContainsInt(a []int, b int) bool {
	for i := 0; i < len(a); i++ {
		if a[i] == b {
			return true
		}
	}
	return false
}

func Contains(cities [][]byte, city []byte) bool {
	for i := 0; i < len(cities); i++ {
		if bytes.Equal(cities[i], city) {
			return true
		}
	}
	return false
}

// Get retrieves the value associated with the given key
func (hm *HashMap) Get(key []byte) *TempData {
	return hm.buckets[hm.hash(key)].Value
}

// Size returns the number of elements in the hashmap
func (hm *HashMap) Size() int {
	return hm.size
}
