package main

import (
	"math"
	"math/rand"
)

var RANDN int = 16

func getIndexQF(x float32, t []float32) uint32 {
    var count uint32 = 0
    for _, val := range t {
        if val < x {
            count++
        }
    }
    return count
}

func computeScore(x, y []float32, t []float32, tabShare1, tabShare2 [][]uint32) (uint32, uint32) {
    var score, mask uint32 = 0, 0
    for i := range x {
        ix := getIndexQF(x[i], t)
        iy := getIndexQF(y[i], t)
        ir := rand.Uint32() % (uint32(math.Pow(2, 16) - 1))
        mask += ir
        score += (tabShare1[ix][iy] + tabShare2[ix][iy]) + ir
    }
    return score, mask
}

func getColumn[T any](board [][]T, columnIndex int) (column []T) {
    column = make([]T, 0)
    for _, row := range board {
        column = append(column, row[columnIndex])
    }
    return column
}

func compIPandIPQ(i int, synSamples [][]float32, t []float32, tabShare1, tabShare2 [][]uint32) (float32, uint32) {
    x := getColumn(synSamples, i)
    y := getColumn(synSamples, i+1)
    var ip float32 = 0
    for i := range x {
        ip += x[i]* y[i]
    }
    ipQ, mask := computeScore(x, y, t, tabShare1, tabShare2)
    return ip, (ipQ - mask)
}
