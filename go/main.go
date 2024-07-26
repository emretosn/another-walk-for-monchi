package main

import (
	"fmt"
	"math"
)


func PearsonCorrelation(x []uint32, y []float32) float32 {
    if len(x) != len(y) {
        return 0
    }

    n := float32(len(x))
    var sumX, sumY, sumXY, sumX2, sumY2 float32

    for i := 0; i < len(x); i++ {
        xi := float32(x[i])
        yi := y[i]
        sumX += xi
        sumY += yi
        sumXY += xi * yi
        sumX2 += xi * xi
        sumY2 += yi * yi
    }

    numerator := n*sumXY - sumX*sumY
    denominator := float32(math.Sqrt(float64((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))))

    if denominator == 0 {
        return 0
    }
    return numerator / denominator
}

func main() {
    // FIRST TEST
    fmt.Println("RANDOM MATRIX DEMO")

    var size uint32 = 5
    selections := []uint32{0, 2, 4}

    MFBR := genRandMFBR(size)
    fmt.Println("MFBR Table")
    printMatrix(MFBR)

    A := selectRows(MFBR, selections)
    fmt.Println("Selections")
    printMatrix(A)

    b := genRandInexes(uint32(len(selections)), uint32(len(MFBR)))
    fmt.Println("Selected Columns", b)
    bmap := genIndexMaps(b, uint32(len(MFBR)))
    fmt.Println("Index Map")
    printMatrix(bmap)

    results := compRowsTimesMasks(A, bmap)
    fmt.Println("Results")
    printMatrix(results)

    // NEXT STEP
    fmt.Println("ADDITIVE SHARE SYNTHETIC")

    numSamples := 2*1000
    idSynSamples := arange(0, numSamples, 2)

    synPath := "./data/Synthetic/syntheticSamples_dimF_512.csv"
    bordersPath := "./lookupTables/Borders/Borders_nB_3_dimF_512.csv"
    tabRandPath := "./lookupTables/Rand/Rand_nB_3_dimF_512.csv"
    tabQMFIPPath := "./lookupTables/MFIP-Rand/MFIPSubRand_nB_3_dimF_512.csv"

    synSamples, err := readCSVToArray(synPath, "[][]float32")
    borders, err := readCSVToArray(bordersPath, "[]float32")
    tabRand, err := readCSVToArray(tabRandPath, "[][]uint32")
    tabQMFIP, err := readCSVToArray(tabQMFIPPath, "[][]uint32")
    if err != nil {
        fmt.Println("Error reading one of the csv files:", err)
    }

    var scoresIP []float32
    var scoresIPQ []uint32

    for _, id := range idSynSamples {
        ip, ipQ := compIPandIPQ(id, synSamples.([][]float32), borders.([]float32), tabQMFIP.([][]uint32), tabRand.([][]uint32))
        scoresIP = append(scoresIP, float32(ip))
        scoresIPQ = append(scoresIPQ, uint32(ipQ))
    }
    correlation := PearsonCorrelation(scoresIPQ, scoresIP)
    fmt.Printf("Pearson Correlation: %.4f\n", correlation)

}
