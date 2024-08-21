package main

import "github.com/tuneinsight/lattigo/v4/rlwe"

func selectRows[T any](tableMFBR [][]T, selections []int) [][]T {
    A := make([][]T, len(selections))
    for i, row := range selections {
        A[i] = tableMFBR[row]
    }
    return A
}

func genFlatIndexMaps(indexes []int64, cols int) []int64 {
    bi := make([]int64, len(indexes)*cols)
    for i, index := range indexes {
        bi[(i*cols)+int(index)] = 1
    }
    return bi
}

// Clear row selection to clear probe mask
func (BIP *BIP_s) compFlatRowsTimesMasks(A []int64, b []uint64) []int64 {
    result := make([]int64, len(A))
    for i := range A {
        result[i] = A[i] * int64(b[i])
    }
    return result
}

/*
    Previous iterations for the row selections
*/
func (BIP *BIP_s) compCFlatRowsTimesMasks(A []*rlwe.Ciphertext, b []uint64) []*rlwe.Ciphertext {
    result := make([]*rlwe.Ciphertext, len(A))
    for i := range A {
        result[i] = BIP.evaluator.MulScalarNew(A[i], b[i])
    }
    return result
}

// For encrypted selection mask
func (BIP *BIP_s) compCCFlatRowsTimesMasks(A, b []*rlwe.Ciphertext) []*rlwe.Ciphertext {
    result := make([]*rlwe.Ciphertext, len(A))
    for i := range A {
        result[i] = BIP.evaluator.MulNew(A[i], b[i])
    }
    return result
}
