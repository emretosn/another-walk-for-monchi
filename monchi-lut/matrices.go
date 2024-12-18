package main

import (
	"encoding/csv"
	"fmt"
    "log"
	"math"
	"math/rand"
	"os"
	"strconv"
)

func genRandMFBR(size int64) [][]int64 {
	X := make([][]int64, size)
	for row := range X {
		X[row] = make([]int64, size)
		for col := range X[row] {
			X[row][col] = int64(rand.Intn(int(math.Pow(float64(size), 2))))
		}
	}
	return X
}

func genRandInexes(size int, maxval int) []int64 {
	b := make([]int64, size)
	for i := range size {
		b[i] = int64(rand.Intn(maxval))
	}
	return b
}

func printMatrix[T any](matrix [][]T) {
	for _, m := range matrix {
		fmt.Println(m)
	}
}

func flatten[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

func readCSVTo2DSlice(filename string) ([][]int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var result [][]int64

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("unable to read CSV file: %v", err)
		}

		var row []int64

		for _, value := range record {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("unable to parse float value: %v", err)
			}
			row = append(row, intVal)
		}
		result = append(result, row)
	}

	return result, nil
}

func readCSVToFloatSlice(filename string) ([]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	record, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("unable to read CSV file: %v", err)
	}

	var result []float64

	for _, value := range record {
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse float value: %v", err)
		}
		result = append(result, floatVal)
	}

	return result, nil
}

func addRecord(writer *csv.Writer, record []string) {
    err := writer.Write(record)
    if err != nil {
        log.Fatalf("Failed to write record to CSV: %s", err)
    }
    writer.Flush()
}
