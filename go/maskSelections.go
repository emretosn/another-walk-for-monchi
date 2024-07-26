package main

//TODO: Encrypt the selectRows with:
/*
import (
    "github.com/tuneinsight/lattigo/v5/schemes/bfv"
)
*/

func selectRows(tableMFBR [][]uint32, selections []uint32) [][]uint32 {
    A := make([][]uint32, len(selections))
    for i, row := range selections {
        A[i] = tableMFBR[row]
    }
    return A
}

func genIndexMaps(indexes []uint32, cols uint32) [][]uint32 {
    bi := make([][]uint32, len(indexes))
    for i, index := range indexes {
        bi[i] = make([]uint32, cols)
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

func compRowsTimesMasks(A [][]uint32, b [][]uint32) [][]uint32 {
    result := make([][]uint32, len(A))
    for i := range A {
        result[i] = make([]uint32, len(A[i]))
        for j := range A[i] {
            result[i][j] = A[i][j] * b[i][j]
        }
    }
    return result
}
