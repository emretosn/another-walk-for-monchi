package main

import "github.com/tuneinsight/lattigo/v4/rlwe"

func selectRows[T any](tableMFBR [][]T, selections []int) [][]T {
    A := make([][]T, len(selections))
    for i, row := range selections {
        A[i] = tableMFBR[row]
    }
    return A
}

func genIndexMaps(indexes []int64, cols int) [][]int64 {
    bi := make([][]int64, len(indexes))
    for i, index := range indexes {
        bi[i] = make([]int64, cols)
        for j := range bi[i] {
            if j == int(index) {
                bi[i][j] = 1
            } else {
                bi[i][j] = 0
            }
        }
    }
    return bi
}

func compRowsTimesMasks(A [][]int64, b [][]int64) [][]int64 {
    result := make([][]int64, len(A))
    for i := range A {
        result[i] = make([]int64, len(A[i]))
        for j := range A[i] {
            result[i][j] = A[i][j] * b[i][j]
        }
    }
    return result
}

func compFlatRowsTimesMasks(A []*rlwe.Ciphertext, b []int64) []*rlwe.Ciphertext {
    result := make([]*rlwe.Ciphertext, len(A))
    for i := range A {
        if b[i] == 0 {
            result[i] = nil
        } else if b[i] == 1 {
            result[i] = A[i]
        }
    }
    return result
}
