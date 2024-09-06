package main

/*
#cgo CFLAGS: -I../funshade/funshade/c
#cgo LDFLAGS: -L../funshade/build -laes -lfss -Wl,-rpath,../funshade/build

#include "aes.h"
#include "fss.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"
    "path/filepath"
    "encoding/csv"

	"github.com/tuneinsight/lattigo/v4/bfv"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/utils"
)

const SEED  = 54321
const NFEAT = 128
const NROWS = 8
const K     = 1
const THETA = 200

const DBSIZE = 200

func main() {
    // SETTING FHE PARAMETERS
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

    crs := []byte("optional")

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

    //READING THE DATA AND TABLE CONVERSION
    mfipPath := "./lookupTables/MFIP/MFIP_nB_3_dimF_128.csv"
    borderPath := "./lookupTables/Borders/Borders_nB_3_dimF_128.csv"
    mfip, err := readCSVTo2DSlice(mfipPath)
    if err != nil {
        log.Fatal(err)
    }
    borders, err := readCSVToFloatSlice(borderPath)
    if err != nil {
        log.Fatal(err)
    }
    Enrollment.mfip = mfip
    Enrollment.borders = borders

    // GETTING THE BIOMETRIC DATA
    bioData := ReadBioData("./data/LFW/")
    //fmt.Println(bioData)

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

    addRecord(writer, []string{"LookupTime", "FSSTime", "MultiplicationTime", "RotationTime", "AdditionTime", "o"})

    var boardingTime time.Duration
    counter := 0

    for i:=0; i<DBSIZE; i++ {
        fmt.Println(i, counter)
        counter = i % len(bioData)

        lfwRefPath := bioData[counter][0]
        lfwProbePath := bioData[counter][1]

        lfwRef, err := readCSVToFloatSlice(lfwRefPath)
        if err != nil {
            log.Fatal(err)
        }
        lfwProbe, err := readCSVToFloatSlice(lfwProbePath)
        if err != nil {
            log.Fatal(err)
        }

        Enrollment.x = lfwProbe
        Enrollment.Y = lfwRef

        // PROBE AND REFERENCE
        lfwRefQ := quantizeFeatures(Enrollment.borders, Enrollment.Y)
        refTemp := refTemplate(lfwRefQ, Enrollment.mfip)
        //fmt.Println("refTemp:", refTemp)

        permutations := genPermutationsConcat(SEED, NFEAT, NROWS)
        permutationsInv := getPermutationsInverse(permutations)

        quantizedProbe := quantizeFeatures(Enrollment.borders , Enrollment.x)
        //fmt.Println("quantizedProbe", quantizedProbe)

        // FSS RANDOMNESS
        P0.r_in, P1.r_in, P0.k, P1.k = FssGenSign(K, THETA)

        BIP.r_in = make([]int32, K)
        for i := range P0.r_in {
            BIP.r_in[i] = P0.r_in[i] + P1.r_in[i]
        }

        // Adding fss randomness to selected columns before ecnryption
        r_values := divideIntoParts(BIP.r_in[0], NFEAT)
        //fmt.Println("BIP.r_in divided:", r_values)

        ctxtSelection := Enrollment.encryptPermutedRefTempSingleCT(r_values, refTemp, permutations)
        BIP.c_selection = ctxtSelection

        permProbeTemp := genPermProbeTemplateFromPermInv(quantizedProbe, permutationsInv, NROWS);
        permProbeTempMask := getPermutedProbeTempMask(permProbeTemp, Enrollment.params.N())
        //fmt.Println("permProbeTempMask:", permProbeTempMask)

        Gate.col_selection = permProbeTempMask

        // TIMING SCORE
        boardingStart := time.Now()
        lookupTime := time.Now()
        result, mulTime, rotTime, addTime := P0.getFinalScoreCT(BIP, Gate.col_selection)

        // Decrypt Score
        encOut := CKSDecrypt(P0.params, PPool, result)
        ptres := bfv.NewPlaintext(params, params.MaxLevel())
        P0.decryptor.Decrypt(encOut, ptres)
        res0 := P0.encoder.DecodeIntNew(ptres)
        //fmt.Println("getFinalScoreCT:", res[:32])

        encOut = CKSDecrypt(P1.params, PPool, result)
        ptres = bfv.NewPlaintext(params, params.MaxLevel())
        P0.decryptor.Decrypt(encOut, ptres)
        res1 := P0.encoder.DecodeIntNew(ptres)

        lookupEnclosed := time.Since(lookupTime)
        fmt.Println("Lookup Time:", lookupEnclosed)

        x_hat0 := make([]int32, 1)
        x_hat0[0] = int32(res0[0])
        //fmt.Println("x_hat:", x_hat)

        x_hat1 := make([]int32, 1)
        x_hat1[0] = int32(res1[0])

        // FSS EVAL
        fssTime := time.Now()

        o_0, err := FssEvalSign(K, false, P0.k, x_hat0)
        if err != nil {
            log.Fatal(err)
        }
        o_1, err := FssEvalSign(K, true, P1.k, x_hat1)
        if err != nil {
            log.Fatal(err)
        }
        o := make([]uint16, len(o_0))
        for i := range o_0 {
            o[i] = o_0[i] + o_1[i]
        }

        fssEnclosed := time.Since(fssTime)
        fmt.Println("FSS Time:", fssEnclosed)

        fmt.Println("o:", o)
        boardingEnd := time.Since(boardingStart)
        boardingTime += boardingEnd

        oS := fmt.Sprintf("%d", o)
        addRecord(writer, []string{lookupEnclosed.String(), fssEnclosed.String(), mulTime.String(), rotTime.String(), addTime.String(), oS})

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

