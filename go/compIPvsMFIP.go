package main

import (
	"math"
	"math/rand"
)

var RANDN int = 16

func getIndexQF(x uint32, t []uint32) uint32 {
    var count uint32 = 0
    for _, val := range t {
        if val < x {
            count++
        }
    }
    return count
}

func computeScore(x, y, t []uint32, tabShare1, tabShare2 [][]uint32) (uint32, uint32) {
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

func compIPandIPQ(i uint32, synSamples, t []uint32, tabShare1, tabShare2 [][]uint32) (float32, uint32) {
    x := synSamples[:i]
    y := synSamples[:i+1]
    var ip float32 = 0
    for i := range x {
        ip += float32(x[i]) * float32(y[i])
    }
    ipQ, mask := computeScore(x, y, t, tabShare1, tabShare2)
    return ip, (ipQ - mask)
}
