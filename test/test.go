package main

import (
    "os"
	"fmt"
    "math"
    "math/rand"
    "encoding/csv"
    "strconv"
    "sync"
	//"math/bits"
    //"math/rand"

	"github.com/tuneinsight/lattigo/v4/bfv"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

func (Enrollment *Enrollment_s) optimizedPlaintextMul(arr []int64) *bfv.PlaintextMul {
    plainMask := bfv.NewPlaintextMul(Enrollment.params, Enrollment.params.MaxLevel())
    Enrollment.encoder.EncodeMul(arr, plainMask)
    return plainMask
}

func (BIP *BIP_s) getFinalScoreCT(Enrollment *Enrollment_s, permRefTempCT *rlwe.Ciphertext, permProbeTempMask []int64) *rlwe.Ciphertext {
    ringDim := BIP.params.N()
    halfRing := float64(ringDim / 2)
    //fmt.Println("halfRing", halfRing)

    permProbeTempMaskPT := Enrollment.optimizedPlaintextMul(permProbeTempMask)
    finalScoreCT := BIP.evaluator.MulNew(permRefTempCT, permProbeTempMaskPT)

    maskedRefTempCT := BIP.evaluator.MulNew(permRefTempCT, permProbeTempMaskPT)

    finalScoreCT = maskedRefTempCT

    //fmt.Println(int(math.Log2(halfRing)))
    for i := 0; i < int(math.Log2(halfRing)); i++ {
		rotation := int(math.Pow(2, float64(i)))

		rotatedCT := BIP.evaluator.RotateColumnsNew(finalScoreCT, rotation)
		BIP.evaluator.Add(finalScoreCT, rotatedCT, finalScoreCT)
	}

	return finalScoreCT
}

func flattenMatrix(matrix [][]int64) []int64 {
	var flattened []int64
	for _, row := range matrix {
		flattened = append(flattened, row...)
	}
	return flattened
}

type BIP_s struct {
	params      bfv.Parameters
	evaluator   bfv.Evaluator
    decryptor   rlwe.Decryptor
}

type Enrollment_s struct {
	params    bfv.Parameters
	encoder   bfv.Encoder
	encryptor rlwe.Encryptor
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

func (Enrollment *Enrollment_s) encryptPermutedRefTempSingleCT(refTemp []int64, permutations []int) *rlwe.Ciphertext{
    ringDim := Enrollment.params.N()
    permRefTemp := make([]int64, ringDim)

    for i := range refTemp {
        permRefTemp[i] = refTemp[permutations[i]]
    }
    ptxt := bfv.NewPlaintext(Enrollment.params, Enrollment.params.MaxLevel())
    Enrollment.encoder.Encode(permRefTemp, ptxt)
    ctxt := bfv.NewCiphertext(Enrollment.params, 1, Enrollment.params.MaxLevel())
    Enrollment.encryptor.Encrypt(ptxt, ctxt)

    return ctxt
}

func getIndexInVect(permutations []int, value int) int {
	for i, v := range permutations {
		if v == value {
			return i
		}
	}
	return -1
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

func getPermutedProbeTempMask(permProbeTemp []int, ringDim int) []int64 {
	nFeat := len(permProbeTemp)
	permProbeTempMask := make([]int64, ringDim)

	var wg sync.WaitGroup
	wg.Add(nFeat)

	for i := 0; i < nFeat; i++ {
		go func(i int) {
			defer wg.Done()
			permProbeTempMask[permProbeTemp[i]] = 1
		}(i)
	}

	wg.Wait()

	return permProbeTempMask
}

const SEED  = 54321
const NFEAT = 128
const NROWS = 8

func main() {
    mfipPath := "../go/lookupTables/MFIP/MFIP_nB_3_dimF_128.csv"
    lfwPath := "../go/data/LFW/Paul_McCartney/0.csv"
    borderPath := "../go/lookupTables/Borders/Borders_nB_3_dimF_128.csv"

    mfip, err := readCSVTo2DSlice(mfipPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }
    lfw, err := readCSVToFloatSlice(lfwPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }
    borders, err := readCSVToFloatSlice(borderPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }

    //fmt.Println("MFIP", mfip)

    //fmt.Println(lfw)
    lfwQ := quantizeFeatures(borders, lfw)
    //fmt.Println("LFWQ", lfwQ)

    refTemp := refTemplate(lfwQ, mfip)
    //fmt.Printf("refTemp %v\n", refTemp)

    permutations := genPermutationsConcat(SEED, NFEAT, NROWS)
    permutationsInv := getPermutationsInverse(permutations)
    fmt.Println("permutations", permutations)

    // Enryption
    paramsDef := bfv.PN13QP218
    paramsDef.T = 0x3ee0001

    paramsDef.LogN = 10
	params, err := bfv.NewParametersFromLiteral(paramsDef)
	if err != nil {
		panic(err)
	}

    kgen := bfv.NewKeyGenerator(params)
    sk, pk := kgen.GenKeyPair()

    rlk := kgen.GenRelinearizationKey(sk, 1)

    halfRing := params.N() / 2
    maxRotation := math.Log2(float64(halfRing))

    rotations := []int{}
    for i := 0; i <= int(maxRotation); i++ {
        rotations = append(rotations, int(math.Pow(2, float64(i))))
    }
    rtks := kgen.GenRotationKeysForRotations(rotations, true, sk)
    evk := rlwe.EvaluationKey{Rlk: rlk, Rtks: rtks}

    Enrollment := &Enrollment_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk)}
    BIP := &BIP_s{params, bfv.NewEvaluator(params, evk), bfv.NewDecryptor(params, sk)}

    ctxt := Enrollment.encryptPermutedRefTempSingleCT(refTemp, permutations)

    ptxt := BIP.decryptor.DecryptNew(ctxt)
    refPerm := Enrollment.encoder.DecodeIntNew(ptxt)

    fmt.Println(refPerm)

    lfwProbPath := "../go/data/LFW/Paul_McCartney/1.csv"

    lfwProb, err := readCSVToFloatSlice(lfwProbPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }
    quantizedProb := quantizeFeatures(borders , lfwProb)
    fmt.Println("quantizedProb", quantizedProb)

    permProbeTemp := genPermProbeTemplateFromPermInv(quantizedProb, permutationsInv, NROWS);
    permProbeTempMask := getPermutedProbeTempMask(permProbeTemp, Enrollment.params.N())
    fmt.Println("permProbeTempMask", permProbeTempMask)
}

