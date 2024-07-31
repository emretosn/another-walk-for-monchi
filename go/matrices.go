package main

import (
	"encoding/csv"
	"errors"
	"fmt"
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

func arange(start, stop, step int) []int {
	var result []int
	for i := start; i < stop; i += step {
		result = append(result, i)
	}
	return result
}

func flatten[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

func readCSVToArray(filename, valtype string) (interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(file)
	stringRows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	switch valtype {
	case "[]float64":
		var float64Array []float64
		for _, value := range stringRows[0] {
			float64Value, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			float64Array = append(float64Array, float64(float64Value))
		}
		return float64Array, nil
	case "[][]int64":
		var int64Matrix [][]int64
		for _, stringRow := range stringRows {
			var int64Row []int64
			for _, value := range stringRow {
				int64Value, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, err
				}
				int64Row = append(int64Row, int64(int64Value))
			}
			int64Matrix = append(int64Matrix, int64Row)
		}
		return int64Matrix, nil
	case "[][]float64":
		var float64Matrix [][]float64
		for _, stringRow := range stringRows {
			var float64Row []float64
			for _, value := range stringRow {
				float64Value, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return nil, err
				}
				float64Row = append(float64Row, float64(float64Value))
			}
			float64Matrix = append(float64Matrix, float64Row)
		}
		return float64Matrix, nil
	default:
		return nil, errors.New("Unexpected type")
	}
}
