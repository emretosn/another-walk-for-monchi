package main

import (
	"encoding/csv"
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

func readCSVToArray(filename string) (interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(file)
	stringRows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
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
}

func quantizeFeatures(borders []float64, unQuantizedFeatures []float64) []int64 {
	numFeat := len(unQuantizedFeatures)
	quantizedFeatures := make([]int64, 0, numFeat)
	lenBorders := len(borders)

	for i := 0; i < numFeat; i++ {
		feature := unQuantizedFeatures[i]
		count := 0

		for count < lenBorders && borders[count] <= feature {
			count++
		}

		quantizedFeatures = append(quantizedFeatures, int64(count))
	}

	return quantizedFeatures
}
