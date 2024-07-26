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


func genRandMFBR(size uint32) [][]uint32 {
    X := make([][]uint32, size)
    for row := range X{
        X[row] = make([]uint32, size)
        for col := range X[row]{
            X[row][col] = uint32(rand.Intn(int(math.Pow(float64(size), 2))))
        }
    }
    return X
}

func genRandInexes(size uint32, maxval uint32) []uint32 {
    b := make([]uint32, size)
    for i := range size {
        b[i] = uint32(rand.Intn(int(maxval)))
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
        var float32Array []float32
        for _, value := range stringRows[0] {
            float32Value, err := strconv.ParseFloat(value, 32)
            if err != nil {
                return nil, err
            }
            float32Array = append(float32Array, float32(float32Value))
        }
        return float32Array, nil
    case "[][]uint32":
        var uint32Matrix [][]uint32
        for _, stringRow := range stringRows {
            var uint32Row []uint32
            for _, value := range stringRow {
                uint32Value, err := strconv.ParseUint(value, 10, 32)
                if err != nil {
                    return nil, err
                }
                uint32Row = append(uint32Row, uint32(uint32Value))
            }
            uint32Matrix = append(uint32Matrix, uint32Row)
        }
        return uint32Matrix, nil
    case "[][]float32":
        var float32Matrix [][]float32
        for _, stringRow := range stringRows {
            var float32Row []float32
            for _, value := range stringRow {
                float32Value, err := strconv.ParseFloat(value, 32)
                if err != nil {
                    return nil, err
                }
                float32Row = append(float32Row, float32(float32Value))
            }
            float32Matrix = append(float32Matrix, float32Row)
        }
        return float32Matrix, nil
    default:
        return nil, errors.New("Unexpected type")
    }
}

