package storage

import (
	"strconv"
	"testing"

	models "github.com/wurt83ow/tinyurl/internal/models"
)

func BenchmarkStorageInsertURL(b *testing.B) {
	var testData []string
	for i := 0; i < 1024; i++ {
		testData = append(testData, strconv.Itoa(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set := MemoryStorage{data: make(map[string]models.DataURL)}
		for _, key := range testData {
			_, err := set.InsertURL(key, models.DataURL{})
			if err != nil {
				b.Fatal("error when inserting element func InsertURL")
			}
		}
	}
}

func BenchmarkStorageInsertUser(b *testing.B) {
	var testData []string
	for i := 0; i < 1024; i++ {
		testData = append(testData, strconv.Itoa(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set := MemoryStorage{users: make(map[string]models.DataUser)}
		for _, key := range testData {
			_, err := set.InsertUser(key, models.DataUser{})
			if err != nil {
				b.Fatal("error when inserting element func InsertURL")
			}
		}
	}
}
func BenchmarkStorageGetURL(b *testing.B) {
	var testData []string
	for i := 0; i < 1024; i++ {
		testData = append(testData, strconv.Itoa(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		set := MemoryStorage{data: make(map[string]models.DataURL)}
		for _, key := range testData {
			_, err := set.InsertURL(key, models.DataURL{})
			if err != nil {
				b.Fatal("error when inserting element func InsertURL")
			}
		}
		b.StartTimer()
		for _, key := range testData {
			nv, _ := set.GetURL(key)

			blackhole := nv
			_ = blackhole

		}
	}
}

func BenchmarkStorageGetUser(b *testing.B) {
	var testData []string
	for i := 0; i < 1024; i++ {
		testData = append(testData, strconv.Itoa(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		set := MemoryStorage{users: make(map[string]models.DataUser)}
		for _, key := range testData {
			_, err := set.InsertUser(key, models.DataUser{})
			if err != nil {
				b.Fatal("error when inserting element func InsertURL")
			}
		}
		b.StartTimer()
		for _, key := range testData {
			nv, _ := set.GetUser(key)

			blackhole := nv
			_ = blackhole

		}
	}
}
