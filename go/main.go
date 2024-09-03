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
	"fmt"
	"log"
	"time"
	"unsafe"

	"github.com/tuneinsight/lattigo/v4/bfv"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/utils"
)

const SEED  = 54321
const NFEAT = 128
const NROWS = 8
const K     = 1
const THETA = 500

func main() {
    //READING THE DATA AND TABLE CONVERSION
    mfipPath := "./lookupTables/MFIP/MFIP_nB_3_dimF_128.csv"
    borderPath := "./lookupTables/Borders/Borders_nB_3_dimF_128.csv"
    //lfwRefPath := "./data/LFW/Paul_McCartney/0.csv"
    lfwRefPath := "./data/LFW/John_Lennon/0.csv"
    lfwProbPath := "./data/LFW/Paul_McCartney/1.csv"

    mfip, err := readCSVTo2DSlice(mfipPath)
    if err != nil {
        log.Fatal(err)
    }
    borders, err := readCSVToFloatSlice(borderPath)
    if err != nil {
        log.Fatal(err)
    }
    lfwRef, err := readCSVToFloatSlice(lfwRefPath)
    if err != nil {
        log.Fatal(err)
    }
    lfwProbe, err := readCSVToFloatSlice(lfwProbPath)
    if err != nil {
        log.Fatal(err)
    }

    paramsDef := bfv.PN13QP218
    paramsDef.LogN = 11

    l := 128
	v_max := 1 << 16 // 12 for 32 bits, normally should be 3 for max value
    N := 1 << paramsDef.LogN
	paramsDef.T = getOptimalT(uint64(l), uint64(v_max), float64(N))

    //fmt.Println("T", bits.Len64(paramsDef.T))

    params, err := bfv.NewParametersFromLiteral(paramsDef)
    if err != nil {
        log.Fatal(err)
    }

    P0 := &Party_s{params, bfv.NewEncoder(params), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
    P1 := &Party_s{params, bfv.NewEncoder(params), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
    PPool := []*Party_s{P0, P1}

    crs := []byte("eurecom")

    P0.sk = bfv.NewKeyGenerator(P0.params).GenSecretKey()
    P0.decryptor = bfv.NewDecryptor(P0.params, P0.sk)
    corr_pnrg, err := utils.NewKeyedPRNG(crs)
    pnrg, err := utils.NewPRNG()
    P0.corr_prng = corr_pnrg
    P0.prng = pnrg

    P1.sk = bfv.NewKeyGenerator(P1.params).GenSecretKey()
	P1.decryptor = bfv.NewDecryptor(P1.params, P1.sk)
    corr_pnrg, err = utils.NewKeyedPRNG(crs)
    pnrg, err = utils.NewPRNG()
    P1.corr_prng = corr_pnrg
    P1.prng = pnrg

    pk := ColPubKeyGen(PPool)
	rlk := ColRelinKeyGen(PPool)
    rtk := ColRotKeyGen(PPool)
	evk := rlwe.EvaluationKey{Rlk: rlk, Rtks: rtk}

    Enrollment := &Enrollment_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk), nil, nil, nil, nil, nil}
	BIP := &BIP_s{params, bfv.NewEvaluator(params, evk), nil, nil}
	Gate := &Gate_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk), nil, nil}

    P0.evaluator =  bfv.NewEvaluator(params, evk)
    P1.evaluator =  bfv.NewEvaluator(params, evk)

    Enrollment.mfip = mfip
    Enrollment.x = lfwProbe
    Enrollment.Y = lfwRef
    Enrollment.borders = borders

    // PROBE AND REFERENCE
    lfwRefQ := quantizeFeatures(Enrollment.borders, Enrollment.Y)
    refTemp := refTemplate(lfwRefQ, Enrollment.mfip)
    //fmt.Println("refTemp:", refTemp)

    permutations := genPermutationsConcat(SEED, NFEAT, NROWS)
    permutationsInv := getPermutationsInverse(permutations)

    quantizedProbe := quantizeFeatures(Enrollment.borders , Enrollment.x)
    //fmt.Println("quantizedProbe", quantizedProbe)

    // FSS RANDOMNESS
    startFSS := time.Now()
    P0.r_in, P1.r_in, P0.k, P1.k = FssGenSign(K, THETA)
    endFSS := time.Now()
    fmt.Println("FSS timing:", endFSS.Sub(startFSS))

    BIP.r_in = make([]int32, K)
    for i := range P0.r_in {
        BIP.r_in[i] = P0.r_in[i] + P1.r_in[i]
    }
    fmt.Printf("P0 r_in : %v\n", P0.r_in[:1])
    fmt.Printf("P1 r_in : %v\n", P1.r_in[:1])
    fmt.Printf("BIP r_in: %v\n", BIP.r_in[:1])

    // Adding fss randomness to selected columns before ecnryption
    r_values := divideIntoParts(BIP.r_in[0], NFEAT)
    fmt.Println("BIP.r_in divided:", r_values)

    ctxtSelection := Enrollment.encryptPermutedRefTempSingleCT(r_values, refTemp, permutations)
    BIP.c_selection = ctxtSelection

    encOut1 := CKSDecrypt(P0.params, PPool, ctxtSelection)
    ptres1 := bfv.NewPlaintext(P0.params, params.MaxLevel())
	P0.decryptor.Decrypt(encOut1, ptres1)
    res1 := P0.encoder.DecodeIntNew(ptres1)
    fmt.Println("ctxtSelection:", res1[:32])

    permProbeTemp := genPermProbeTemplateFromPermInv(quantizedProbe, permutationsInv, NROWS);
    permProbeTempMask := getPermutedProbeTempMask(permProbeTemp, Enrollment.params.N())
    //fmt.Println("permProbeTempMask:", permProbeTempMask)

    Gate.col_selection = permProbeTempMask

    // TIMING SCORE
    start := time.Now()

    result := P0.getFinalScoreCT(BIP, Gate.col_selection)

    // Decrypt Score
    encOut := CKSDecrypt(P0.params, PPool, result)
    ptres := bfv.NewPlaintext(params, params.MaxLevel())
	P0.decryptor.Decrypt(encOut, ptres)
    res := P0.encoder.DecodeIntNew(ptres)
    fmt.Println("getFinalScoreCT:", res[:32])

    x_hat := make([]int32, len(res))
    for i, v := range res {
        x_hat[i] = int32(v)
    }

    // FSS EVAL
    o_0, err := FssEvalSign(K, false, P0.k, x_hat[:K])
    if err != nil {
        log.Fatal(err)
    }
    o_1, err := FssEvalSign(K, true, P1.k, x_hat[:K])
    if err != nil {
        log.Fatal(err)
    }
    o := make([]uint16, len(o_0))
    for i := range o_0 {
        o[i] = o_0[i] + o_1[i]
    }

    fmt.Println("o  :", o)

    end := time.Now()
    fmt.Println("time", end.Sub(start))
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
