package main

import (
    "os"
	"fmt"
	"math"
	"math/rand"
    "encoding/csv"
    "strconv"
)

// Creating random indexes and matrices for demo

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
    for i := range matrix {
        for j := range matrix[i] {
            fmt.Print(matrix[i][j], " ")
        }
        fmt.Print("\n")
    }
}

func arange(start, stop, step int) []uint32 {
	var result []uint32
	for i := start; i < stop; i += step {
		result = append(result, uint32(i))
	}
	return result
}

func readCSVToUint32Array1d(filename string) ([]uint32, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := csv.NewReader(file)

    stringRows, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    var uint32Array []uint32
    for _, value := range stringRows[0] {
        uint32Value, err := strconv.ParseUint(value, 10, 32)
        if err != nil {
            return nil, err
        }
        uint32Array = append(uint32Array, uint32(uint32Value))
    }
    return uint32Array, nil
}

func readCSVToUint32Array2d(filename string) ([][]uint32, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	stringRows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

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
}

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

    numSamples := 2*100000
    idSynSamples := arange(0, numSamples, 2)

    synPath := "./data/Synthetic/syntheticSamples_dimF_512.csv"
    bordersPath := "./lookupTables/Borders/Borders_nB_3_dimF_512.csv"
    tabRandPath := "./lookupTables/Rand/Rand_nB_3_dimF_512.csv"
    tabQMFIPPath := "./lookupTables/MFIP-Rand/MFIPSubRand_nB_3_dimF_512.csv"

    synSamples, err := readCSVToUint32Array1d(synPath)
    borders, err := readCSVToUint32Array1d(bordersPath)
    tabRand, err := readCSVToUint32Array2d(tabRandPath)
    tabQMFIP, err := readCSVToUint32Array2d(tabQMFIPPath)
    if err != nil {
        fmt.Println("Error reading one of the csv files:", err)
    }

    var scoresIP []float32
    var scoresIPQ []uint32

    for _, id := range idSynSamples {
        ip, ipQ := compIPandIPQ(id, synSamples, borders, tabQMFIP, tabRand)
        fmt.Printf("ip: %v, ipQ: %v \n", ip, ipQ)
        scoresIP = append(scoresIP, float32(ip))
        scoresIPQ = append(scoresIPQ, uint32(ipQ))
    }
    correlation := PearsonCorrelation(scoresIPQ, scoresIP)
    fmt.Printf("Pearson Correlation: %.4f\n", correlation)

}
