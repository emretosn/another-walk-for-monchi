package main

/*
#cgo CFLAGS: -I./funshade/funshade/c
#cgo LDFLAGS: -L./funshade/build -laes -lfss

#include "aes.h"
#include "fss.h"
#include <stdlib.h>
*/
import "C"

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
    "unsafe"
    "log"
    "path/filepath"
)

func flattenMatrix(matrix [][]int64) []int64 {
	var flattened []int64
	for _, row := range matrix {
		flattened = append(flattened, row...)
	}
	return flattened
}

func readCSVTo2DSlice(filename string) ([][]int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var result [][]int64

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("unable to read CSV file: %v", err)
		}

		var row []int64

		for _, value := range record {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("unable to parse float value: %v", err)
			}
			row = append(row, intVal)
		}
		result = append(result, row)
	}

	return result, nil
}

func readCSVToFloatSlice(filename string) ([]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	record, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("unable to read CSV file: %v", err)
	}

	var result []float64

	for _, value := range record {
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse float value: %v", err)
		}
		result = append(result, floatVal)
	}

	return result, nil
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

func refTemplate(sampleQ []int64, mfbrTab [][]int64) []int64 {
	var refTempPacked []int64
	for _, i := range sampleQ {
		refTemp := mfbrTab[i]
		refTempPacked = append(refTempPacked, refTemp...)
	}
	return refTempPacked
}

func genVectOfInt(begin, length int) []int {
	vect := make([]int, length)
	for i := range vect {
		vect[i] = begin + i
	}
	return vect
}
func getIndexInVect(permutations []int, value int) int {
	for i, v := range permutations {
		if v == value {
			return i
		}
	}
	return -1
}
func vectIntPermutation(seed int64, begin, length int) []int {
	r := rand.New(rand.NewSource(seed))
	seed = r.Int63()

	permuted := genVectOfInt(begin, length)

	rand.Shuffle(len(permuted), func(i, j int) {
		permuted[i], permuted[j] = permuted[j], permuted[i]
	})

	return permuted
}

func addSameValToVector(vect []int, val int) []int {
	for i := range vect {
		vect[i] += val
	}
	return vect
}

func genPermutationsConcat(seed int64, nPerm, lenPerm int) []int {
	permutations := make([]int, 0, nPerm*lenPerm)

	for i := 0; i < nPerm; i++ {
		perm := vectIntPermutation(seed+int64(i), 0, lenPerm)
		perm = addSameValToVector(perm, i*lenPerm)
		permutations = append(permutations, perm...)
	}

	return permutations
}
func getPermutationsInverse(permutations []int) []int {
	nLen := len(permutations)
	permutationsInverse := make([]int, nLen)

	var wg sync.WaitGroup
	wg.Add(nLen)

	for i := 0; i < nLen; i++ {
		go func(i int) {
			defer wg.Done()
			permutationsInverse[i] = getIndexInVect(permutations, i)
		}(i)
	}

	wg.Wait()

	return permutationsInverse
}
func genRefTempFromPerm(refTemp []int64, permutations []int, r_values []int32) []int64 {
	dim := len(refTemp)
	permRefTemp := make([]int64, dim)

    j := -1
	for i := 0; i < dim; i++ {
        if i%8 == 0 {
            j++
        }
		permRefTemp[i] = refTemp[permutations[i]] + int64(r_values[j])
	}
	return permRefTemp
}

func genPermProbeTemplateFromPermInv(quantizedProb []int64, permutationsInverse []int, lenRow int) []int {
	dim := len(quantizedProb)
	permProbeTemp := make([]int, dim)

	var wg sync.WaitGroup
	wg.Add(dim)

	for i := 0; i < dim; i++ {
		go func(i int) {
			defer wg.Done()
			val := int(quantizedProb[i]) + i*lenRow
			permProbeTemp[i] = permutationsInverse[val]
		}(i)
	}

	wg.Wait()
	return permProbeTemp
}

// create a function that take a vector and create 2 additive shares of it
func createAdditiveShares(vector []int64) ([]int64, []int64) {
	n := len(vector)
	share1 := make([]int64, n)
	share2 := make([]int64, n)
    j := -1
	for i := 0; i < n; i++ {
		if i%8 == 0 {
			j++
		}
		// generate a random number in the range of int32 and cast it to int64
		share1[i] = int64(rand.Int31())
		share2[i] = vector[i] - share1[i]
	}
	// assert that the sum of the shares is equal to the original vector

	return share1, share2
}

// create a lookup function that takes a flatten table and a vector of row indicies and return the corresponding rows
func lookupTable(table []int64, indices []int) int64 {
	n := len(indices)
	result := int64(0)
	for i := 0; i < n; i++ {
		result += table[indices[i]]
	}
	return result
}

const SEED = 54321
const NFEAT = 128
const NROWS = 8

const K     = 1
const THETA = 200

const DBSIZE = 1000000

func main() {
	mfipPath := "../go/lookupTables/MFIP/MFIP_nB_3_dimF_128.csv"
    borderPath := "../go/lookupTables/Borders/Borders_nB_3_dimF_128.csv"
    mfip, err := readCSVTo2DSlice(mfipPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }
    borders, err := readCSVToFloatSlice(borderPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }

    // GETTING THE BIOMETRIC DATA
    bioData := ReadBioData("../go/data/LFW/")

    // To save test results
    err = os.MkdirAll("results", os.ModePerm)
    if err != nil {
        log.Fatalf("Failed to create directory: %s", err)
    }
    filename := fmt.Sprintf("%d_results.csv", DBSIZE)
    filePath := filepath.Join("./results", filename)
    if _, err := os.Stat(filePath); err == nil {
        err = os.Remove(filePath)
        if err != nil {
            log.Fatalf("Failed to delete existing file: %s", err)
        }
    }
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("Failed to open file: %s", err)
    }
    defer file.Close()
    writer := csv.NewWriter(file)
    defer writer.Flush()

    addRecord(writer, []string{"LookupTime", "FSSTime", "o"})

    var boardingTime time.Duration
    counter := 0
    for i:=0; i<DBSIZE; i++ {
        fmt.Println(i, counter)
        counter = i % len(bioData)

        lfwRefPath := bioData[counter][0]
        lfwProbePath := bioData[counter][1]

        reference, err := readCSVToFloatSlice(lfwRefPath)
        if err != nil {
            fmt.Println(fmt.Errorf(err.Error()))
        }
        live, err := readCSVToFloatSlice(lfwProbePath)
        if err != nil {
            fmt.Println(fmt.Errorf(err.Error()))
        }

        // fmt.Println("MFIP", mfip)
        r0_in, r1_in, k0, k1 := FssGenSign(K, THETA)
        r_in := make([]int32, K)
        for i := range r_in {
            r_in[i] = r0_in[i] + r1_in[i]
        }
        r_values := divideIntoParts(r_in[0], NFEAT)

        //fmt.Println(r0_in)
        //fmt.Println(r1_in)

        //fmt.Println(lfw)
        referenceQ := quantizeFeatures(borders, reference)
        // fmt.Println("referenceQ", referenceQ)

        liveQ := quantizeFeatures(borders, live)
        //fmt.Println("liveQ", liveQ)
        ////print len of liveQ
        // take liveQ and multiply each value by
        // fmt.Println("len(liveQ)", len(liveQ))

        refTemp := refTemplate(referenceQ, mfip)
        //fmt.Printf("refTemp %v \n len(refTemp) %v \n", refTemp, len(refTemp))

        permutations := genPermutationsConcat(SEED, NFEAT, NROWS)

        permRefTemp := genRefTempFromPerm(refTemp, permutations, r_values)
        // fmt.Println("permRefTemp", permRefTemp)
        permutationsInv := getPermutationsInverse(permutations)

        permProbeTemp := genPermProbeTemplateFromPermInv(liveQ, permutationsInv, NROWS)
        // fmt.Println("permProbeTemp", permProbeTemp)

        //Create additive shares of the refTemp
        share1, share2 := createAdditiveShares(permRefTemp)
        // fmt.Println("share1", share1)
        // fmt.Println("share2", share2)

        //measure the time it takes to perform the lookups
        boardingStart := time.Now()
        start := time.Now()
        maskedSore1 := lookupTable(share1, permProbeTemp)
        maskedScore2 := lookupTable(share2, permProbeTemp)
        // fmt.Println("lookupTable1", maskedSore1)
        // fmt.Println("lookupTable2", maskedScore2)
        result := maskedSore1 + maskedScore2
        end := time.Now()
        lookupT := end.Sub(start)
        fmt.Println("Lookup Time:", lookupT)
        //fmt.Println("result", result)

        res := make([]int32, 1)
        res[0] = int32(result)

        // FSS EVAL
        s := time.Now()
        o_0, err := FssEvalSign(K, false, k0, res)
        if err != nil {
            log.Fatal(err)
        }
        o_1, err := FssEvalSign(K, true, k1, res)
        if err != nil {
            log.Fatal(err)
        }
        o := make([]uint16, len(o_0))
        for i := range o_0 {
            o[i] = o_0[i] + o_1[i]
        }
        e := time.Now()
        fssT := e.Sub(s)
        fmt.Println("FSS Time:", fssT)

        fmt.Println("o:", o)

        boardingEnd := time.Since(boardingStart)
        boardingTime += boardingEnd

        oS := fmt.Sprintf("%d", o)
        addRecord(writer, []string{lookupT.String(), fssT.String(), oS})

        //resultClear := lookupTable(permRefTemp, permProbeTemp)
        //fmt.Println("resultClear", resultClear)
    }
    fmt.Println("Total Boarding Time:", boardingTime)
    addRecord(writer, []string{boardingTime.String()})
}

func FssGenSign(K int32, theta uint16) ([]int32, []int32, []byte, []byte) {
    r_in0 := make([]int32, K)
    r_in1 := make([]int32, K)
    k0 := make([]byte, K*C.KEY_LEN)
    k1 := make([]byte, K*C.KEY_LEN)

    r_in0Ptr := (*C.uint16_t)(unsafe.Pointer(&r_in0[0]))
    r_in1Ptr := (*C.uint16_t)(unsafe.Pointer(&r_in1[0]))
	k0Ptr := (*C.uint8_t)(unsafe.Pointer(&k0[0]))
	k1Ptr := (*C.uint8_t)(unsafe.Pointer(&k1[0]))

    C.SIGN_gen_batch(C.size_t(K), C.uint16_t(theta), r_in0Ptr, r_in1Ptr, k0Ptr, k1Ptr)

    return r_in0, r_in1, k0, k1
}

func FssEvalSign(K int32, j bool, k_j []byte, x_hat []int32) ([]uint16, error) {
    if len(x_hat) != int(K) {
		return nil, fmt.Errorf("<FssEvalSign error> x_hat shares must be of length %d", K)
	}
	if len(k_j) != int(K*C.KEY_LEN) {
		return nil, fmt.Errorf("<FssEvalSign error> FSS keys k_j must be of length %d", K*C.KEY_LEN)
	}

	o_j := make([]uint16, K)

	k_jPtr := (*C.uint8_t)(unsafe.Pointer(&k_j[0]))
	x_hatPtr := (*C.uint16_t)(unsafe.Pointer(&x_hat[0]))
	o_jPtr := (*C.uint16_t)(unsafe.Pointer(&o_j[0]))

	C.SIGN_eval_batch(C.size_t(K), C.bool(j), k_jPtr, x_hatPtr, o_jPtr)

	return o_j, nil
}

func divideIntoParts(value int32, d int) []int32 {
	parts := make([]int32, d)

	if d == 1 {
		parts[0] = value
		return parts
	}

    // FIND A WAY TO INCORPORATE THE REMAINING
	r := rand.New(rand.NewSource(42))
	remaining := value
	for i := 0; i < d-1; i++ {
		// Generate a random number between 0 and remaining value
		parts[i] = r.Int31() % (1 << 16)
		remaining -= parts[i]
	}

	parts[d-1] = remaining

	rand.Shuffle(d, func(i, j int) { parts[i], parts[j] = parts[j], parts[i] })

	return parts
}

// Reads files with at least two photos to a map
func ReadBioData(path string) map[int][]string {
    bioData := make(map[int][]string, 0)
    i := 0

    items, err := os.ReadDir(path)
    if err != nil {
        log.Fatal(err)
    }
    for _, item := range items {
        if item.IsDir() {
            subdirPath := filepath.Join(path, item.Name())
            subitems, err := os.ReadDir(subdirPath)
            if err != nil {
                log.Fatal(err)
            }
            for _, sitem := range subitems {
                if sitem.Name() == "1.csv" {
                    dataRefPath := filepath.Join(subdirPath, "0.csv")
                    dataProbePath := filepath.Join(subdirPath, sitem.Name())
                    bioData[i] = []string{dataRefPath, dataProbePath}
                    i++
                }
            }
        }
    }
    return bioData
}

func addRecord(writer *csv.Writer, record []string) {
    err := writer.Write(record)
    if err != nil {
        log.Fatalf("Failed to write record to CSV: %s", err)
    }
    writer.Flush() // Ensure that the record is written to the file
}

