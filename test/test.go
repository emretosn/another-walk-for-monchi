package main

import (
	"fmt"
    "math"
	"math/bits"
    "math/rand"
    "time"

	"github.com/tuneinsight/lattigo/v4/bfv"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

func (BIP *BIP_s) getFinalScoreCT(Enrollment *Enrollment_s, permRefTempCT *rlwe.Ciphertext, permProbeTempMask []int64) *rlwe.Ciphertext {
    ringDim := BIP.params.N()
    halfRing := float64(ringDim / 2)
    //fmt.Println(ringDim)

    permProbeTempMaskPT := bfv.NewPlaintext(BIP.params, BIP.params.MaxLevel())
    Enrollment.encoder.Encode(permProbeTempMask, permProbeTempMaskPT)
    maskedRefTempCT := BIP.evaluator.MulNew(permRefTempCT, permProbeTempMaskPT)

    finalScoreCT := maskedRefTempCT

    //fmt.Println("finalScoreCT")
    //finalScorePT := BIP.decryptor.DecryptNew(finalScoreCT)
    //finalScoreVec := Enrollment.encoder.DecodeUintNew(finalScorePT)
    //fmt.Println(len(finalScoreVec))
    //fmt.Println(finalScoreVec[:16])

    for i := 0; i < int(math.Log2(halfRing)); i++ {
		rotation := int(math.Pow(2, float64(i)))
        //fmt.Println(rotation)

		rotatedCT := BIP.evaluator.RotateColumnsNew(finalScoreCT, rotation)
		BIP.evaluator.Add(finalScoreCT, rotatedCT, finalScoreCT)

        //fmt.Println("finalScorerotCT")
        //finalScorerotPT := BIP.decryptor.DecryptNew(rotatedCT)
        //finalScorerotVec := Enrollment.encoder.DecodeUintNew(finalScorerotPT)
        //fmt.Println(finalScorerotVec[:16])

        //fmt.Println("finalScoreCT")
        //finalScorePT = BIP.decryptor.DecryptNew(finalScoreCT)
        //finalScoreVec = Enrollment.encoder.DecodeUintNew(finalScorePT)
        //fmt.Println(finalScoreVec[:16])
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

func main() {
    ///////////////////////
    size := 1 << 3

	// Seed the random generator
	rand.Seed(time.Now().UnixNano())

	// Generate the random list u
	u := make([]int, size)
	for i := 0; i < size; i++ {
		u[i] = rand.Intn(size)
	}

	// Create a 2D slice for the one-hot encoding
	oneHot := make([][]int64, 3)
	for i := range oneHot {
		oneHot[i] = make([]int64, size)
	}

	// Set the appropriate positions to 1
	for i := 0; i < 3; i++ {
        for j := 0; j < size; j++ {
            oneHot[i][u[j]] = 1
        }
	}

	// Flatten the one-hot encoded matrix
	permProbeTempMask := flattenMatrix(oneHot)

    ///////////////////////

    const sliceSize = (1 << 3) * 3
	const maxValue = 1 << 3

	// Create a slice with the required size
	permRefTempVec := make([]int64, sliceSize)

	// Populate the slice with random permRefTempVec up to maxValue
	for i := 0; i < sliceSize; i++ {
		permRefTempVec[i] = int64(rand.Intn(maxValue))
	}

    ///////////////////////

	paramsDef := bfv.PN13QP218
    paramsDef.T = 0x3ee0001

    // MinLogN in rlwe/params.go defined as 4 and Max 17
    //paramsDef.LogN = 13
	params, err := bfv.NewParametersFromLiteral(paramsDef)
	if err != nil {
		panic(err)
	}

    kgen := bfv.NewKeyGenerator(params)
    sk, pk := kgen.GenKeyPair()

    rlk := kgen.GenRelinearizationKey(sk, 1)

    halfRing := params.N() / 2
    maxRotation := bits.Len64(uint64(halfRing)) - 1

    rotations := []int{}
    for i := 0; i <= maxRotation; i++ {
        rotations = append(rotations, int(math.Pow(2, float64(i))))
    }
    rtks := kgen.GenRotationKeysForRotations(rotations, true, sk)
    evk := rlwe.EvaluationKey{Rlk: rlk, Rtks: rtks}


    Enrollment := &Enrollment_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk)}
    BIP := &BIP_s{params, bfv.NewEvaluator(params, evk), bfv.NewDecryptor(params, sk)}

    permRefTempPT := Enrollment.encoder.EncodeNew(permRefTempVec, params.MaxLevel())
	permRefTempCT := Enrollment.encryptor.EncryptNew(permRefTempPT)

    permDecPT := BIP.decryptor.DecryptNew(permRefTempCT)
    permDec := Enrollment.encoder.DecodeUintNew(permDecPT)
    fmt.Println("Decoded Ref       :", permDec[:100])

    finalScoreCT := BIP.getFinalScoreCT(Enrollment, permRefTempCT, permProbeTempMask)

    finalScorePT := BIP.decryptor.DecryptNew(finalScoreCT)
    finalScoreVec := Enrollment.encoder.DecodeUintNew(finalScorePT)

    fmt.Println("Original Vector   :", permRefTempVec)
	fmt.Println("Mask Vector       :", permProbeTempMask)
    fmt.Printf("Final Score Vector: %v, len: %v\n", finalScoreVec[:16], len(finalScoreVec))
}

