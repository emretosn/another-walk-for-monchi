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
const K     = 1//1024
const THETA = 1234

func main() {
    //READING THE DATA AND TABLE CONVERSION
    mfipPath := "./lookupTables/MFIP/MFIP_nB_3_dimF_128.csv"
    borderPath := "./lookupTables/Borders/Borders_nB_3_dimF_128.csv"
    lfwPath := "./data/LFW/Paul_McCartney/0.csv"
    //lfwPath := "./data/LFW/John_Lennon/0.csv"
    lfwProbPath := "./data/LFW/Paul_McCartney/1.csv"

    mfip, err := readCSVTo2DSlice(mfipPath)
    if err != nil {
        log.Fatal(err)
    }
    borders, err := readCSVToFloatSlice(borderPath)
    if err != nil {
        log.Fatal(err)
    }
    lfw, err := readCSVToFloatSlice(lfwPath)
    if err != nil {
        log.Fatal(err)
    }
    lfwProb, err := readCSVToFloatSlice(lfwProbPath)
    if err != nil {
        log.Fatal(err)
    }

    //MULTI PARTY MULTI BIP HE
    paramsDef := bfv.PN13QP218
    // Set the propper T value instead of a default later
    paramsDef.T = 0x3ee0001
    // Setting Correct N
    paramsDef.LogN = 11

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

    Enrollment := &Enrollment_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk), nil, nil}
	BIP := &BIP_s{params, bfv.NewEvaluator(params, evk), nil, nil}
	Gate := &Gate_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk), nil, nil}

    P0.evaluator =  bfv.NewEvaluator(params, evk)
    P1.evaluator =  bfv.NewEvaluator(params, evk)

    Enrollment.Y = mfip

    // REFERENCE AND PROBE
    lfwQ := quantizeFeatures(borders, lfw)
    Gate.x = lfwQ

    refTemp := refTemplate(Gate.x, Enrollment.Y)
    //fmt.Println("refTemp:", refTemp)

    permutations := genPermutationsConcat(SEED, NFEAT, NROWS)
    permutationsInv := getPermutationsInverse(permutations)

    quantizedProb := quantizeFeatures(borders , lfwProb)
    //fmt.Println("quantizedProb", quantizedProb)

    // THE FSS RANDOMNESS
    startFSS := time.Now()
    P0.r_in, P1.r_in, P0.k, P1.k = FssGenSign(K, THETA)
    endFSS := time.Now()
    fmt.Println("FSS timing:", endFSS.Sub(startFSS))

    r_in := make([]int32, K)
    // Addition of the modulus operation (?)
    for i := range r_in {
        r_in[i] = P0.r_in[i] + P1.r_in[i]
    }
    fmt.Println(r_in)
    BIP.r_in = make([]int64, len(r_in))
    for i, v := range r_in {
        BIP.r_in[i] = int64(v)
    }
    fmt.Println(BIP.r_in)

    //fmt.Println("P0 r_in", P0.r_in)
    //fmt.Println("P1 r_in", P1.r_in)
    //fmt.Println("BIP r_in", BIP.r_in)

    ctxtSelection := Enrollment.encryptPermutedRefTempSingleCT(refTemp, permutations)

    ptxtAdd := bfv.NewPlaintext(Enrollment.params, Enrollment.params.MaxLevel())
    Enrollment.encoder.Encode(BIP.r_in, ptxtAdd)
    BIP.c_selection = BIP.evaluator.AddNew(ctxtSelection, ptxtAdd)

    permProbeTemp := genPermProbeTemplateFromPermInv(quantizedProb, permutationsInv, NROWS);
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
    o := make([]int32, len(o_0))
    for i := range o_0 {
        o[i] = o_0[i] + o_1[i]
    }

    fmt.Println("Result:", x_hat)

    end := time.Now()
    fmt.Println("time", end.Sub(start))
}

func FssGenSign(K int32, theta int32) ([]int32, []int32, []byte, []byte) {
    r_in0 := make([]int32, K)
    r_in1 := make([]int32, K)
    k0 := make([]byte, K*C.KEY_LEN)
    k1 := make([]byte, K*C.KEY_LEN)

    r_in0Ptr := (*C.int32_t)(unsafe.Pointer(&r_in0[0]))
    r_in1Ptr := (*C.int32_t)(unsafe.Pointer(&r_in1[0]))
	k0Ptr := (*C.uint8_t)(unsafe.Pointer(&k0[0]))
	k1Ptr := (*C.uint8_t)(unsafe.Pointer(&k1[0]))

    C.SIGN_gen_batch(C.size_t(K), C.int(theta), r_in0Ptr, r_in1Ptr, k0Ptr, k1Ptr)

    return r_in0, r_in1, k0, k1
}

func FssEvalSign(K int32, j bool, k_j []byte, x_hat []int32) ([]int32, error) {
    if len(x_hat) != int(K) {
		return nil, fmt.Errorf("<FssEvalSign error> x_hat shares must be of length %d", K)
	}
	if len(k_j) != int(K*C.KEY_LEN) {
		return nil, fmt.Errorf("<FssEvalSign error> FSS keys k_j must be of length %d", K*C.KEY_LEN)
	}

	o_j := make([]int32, K)

	k_jPtr := (*C.uint8_t)(unsafe.Pointer(&k_j[0]))
	x_hatPtr := (*C.int32_t)(unsafe.Pointer(&x_hat[0]))
	o_jPtr := (*C.int32_t)(unsafe.Pointer(&o_j[0]))

	C.SIGN_eval_batch(C.size_t(K), C.bool(j), k_jPtr, x_hatPtr, o_jPtr)

	return o_j, nil
}
