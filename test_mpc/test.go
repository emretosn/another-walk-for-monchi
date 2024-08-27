package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	//"math/bits"
	//"math/rand"
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
func genRefTempFromPerm(refTemp []int64, permutations []int) []int64 {
	dim := len(refTemp)
	permRefTemp := make([]int64, dim)
	for i := 0; i < dim; i++ {
		permRefTemp[i] = refTemp[permutations[i]]
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
	for i := 0; i < n; i++ {
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

func main() {
	mfipPath := "../go/lookupTables/MFIP/MFIP_nB_3_dimF_128.csv"
	refPath := "../go/data/LFW/Paul_McCartney/0.csv"
	livePath := "../go/data/LFW/Paul_McCartney/1.csv"
	borderPath := "../go/lookupTables/Borders/Borders_nB_3_dimF_128.csv"

	mfip, err := readCSVTo2DSlice(mfipPath)
	if err != nil {
		fmt.Println(fmt.Errorf(err.Error()))
	}
	reference, err := readCSVToFloatSlice(refPath)
	if err != nil {
		fmt.Println(fmt.Errorf(err.Error()))
	}
	live, err := readCSVToFloatSlice(livePath)
	if err != nil {
		fmt.Println(fmt.Errorf(err.Error()))
	}
	borders, err := readCSVToFloatSlice(borderPath)
	if err != nil {
		fmt.Println(fmt.Errorf(err.Error()))
	}

	// fmt.Println("MFIP", mfip)

	//fmt.Println(lfw)
	referenceQ := quantizeFeatures(borders, reference)
	// fmt.Println("referenceQ", referenceQ)

	liveQ := quantizeFeatures(borders, live)
	// fmt.Println("liveQ", liveQ)
	//print len of liveQ
	// fmt.Println("len(liveQ)", len(liveQ))

	refTemp := refTemplate(referenceQ, mfip)
	fmt.Printf("refTemp %v \n len(refTemp) %v \n", refTemp, len(refTemp))

	permutations := genPermutationsConcat(SEED, NFEAT, NROWS)

	permRefTemp := genRefTempFromPerm(refTemp, permutations)
	// fmt.Println("permRefTemp", permRefTemp)
	permutationsInv := getPermutationsInverse(permutations)

	permProbeTemp := genPermProbeTemplateFromPermInv(liveQ, permutationsInv, NROWS)
	// fmt.Println("permProbeTemp", permProbeTemp)

	//Create additive shares of the refTemp
	share1, share2 := createAdditiveShares(permRefTemp)
	// fmt.Println("share1", share1)
	// fmt.Println("share2", share2)

	maskedSore1 := lookupTable(share1, permProbeTemp)
	maskedScore2 := lookupTable(share2, permProbeTemp)
	// fmt.Println("lookupTable1", maskedSore1)
	// fmt.Println("lookupTable2", maskedScore2)
	result := maskedSore1 + maskedScore2
	fmt.Println("result", result)

	resultClear := lookupTable(permRefTemp, permProbeTemp)
	fmt.Println("resultClear", resultClear)
}
