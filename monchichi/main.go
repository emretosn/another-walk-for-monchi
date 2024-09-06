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
	"encoding/csv"
	"fmt"
	"os"
	"time"
    "unsafe"
    "log"
    "path/filepath"
)

const SEED = 54321
const NFEAT = 128
const NROWS = 8

const K     = 1
const THETA = 200

const DBSIZE = 200

func main() {
	mfipPath := "../monchi-lut/lookupTables/MFIP/MFIP_nB_3_dimF_128.csv"
    borderPath := "../monchi-lut/lookupTables/Borders/Borders_nB_3_dimF_128.csv"
    mfip, err := readCSVTo2DSlice(mfipPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }
    borders, err := readCSVToFloatSlice(borderPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }

    // GETTING THE BIOMETRIC DATA
    bioData := ReadBioData("../monchi-lut/data/LFW/")

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
