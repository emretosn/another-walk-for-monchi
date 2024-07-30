package main

import (
    "os"
    "fmt"
    "errors"
    "math"
    "math/rand"
    "encoding/csv"
    "strconv"
)


func genRandMFBR(size uint64) [][]uint64 {
    X := make([][]uint64, size)
    for row := range X{
        X[row] = make([]uint64, size)
        for col := range X[row]{
            X[row][col] = uint64(rand.Intn(int(math.Pow(float64(size), 2))))
        }
    }
    return X
}

func genRandInexes(size int, maxval int) []uint64 {
    b := make([]uint64, size)
    for i := range size {
        b[i] = uint64(rand.Intn(maxval))
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
    case "[]float32":
        var float64Array []float32
        for _, value := range stringRows[0] {
            float64Value, err := strconv.ParseFloat(value, 64)
            if err != nil {
                return nil, err
            }
            float64Array = append(float64Array, float32(float64Value))
        }
        return float64Array, nil
    case "[][]uint64":
        var uint64Matrix [][]uint64
        for _, stringRow := range stringRows {
            var uint32Row []uint64
            for _, value := range stringRow {
                uint32Value, err := strconv.ParseUint(value, 10, 64)
                if err != nil {
                    return nil, err
                }
                uint32Row = append(uint32Row, uint64(uint32Value))
            }
            uint64Matrix = append(uint64Matrix, uint32Row)
        }
        return uint64Matrix, nil
    case "[][]float32":
        var float64Matrix [][]float32
        for _, stringRow := range stringRows {
            var float32Row []float32
            for _, value := range stringRow {
                float64Value, err := strconv.ParseFloat(value, 64)
                if err != nil {
                    return nil, err
                }
                float32Row = append(float32Row, float32(float64Value))
            }
            float64Matrix = append(float64Matrix, float32Row)
        }
        return float64Matrix, nil
    default:
        return nil, errors.New("Unexpected type")
    }
}

