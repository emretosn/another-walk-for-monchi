package main

import (
	"fmt"
	"time"

	"github.com/tuneinsight/lattigo/v4/bfv"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/utils"
)

const SEED  = 54321
const NFEAT = 128
const NROWS = 8

func main() {
    //READING THE DATA AND TABLE CONVERSION
    mfipPath := "./lookupTables/MFIP/MFIP_nB_3_dimF_128.csv"
    borderPath := "./lookupTables/Borders/Borders_nB_3_dimF_128.csv"
    lfwPath := "./data/LFW/Paul_McCartney/0.csv"
    //lfwPath := "./data/LFW/John_Lennon/0.csv"
    lfwProbPath := "./data/LFW/Paul_McCartney/1.csv"

    mfip, err := readCSVTo2DSlice(mfipPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }
    borders, err := readCSVToFloatSlice(borderPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }
    lfw, err := readCSVToFloatSlice(lfwPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }
    lfwProb, err := readCSVToFloatSlice(lfwProbPath)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }

    //MULTI PARTY MULTI BIP HE
    paramsDef := bfv.PN13QP218
    // Set the propper T value instead of a default later
    paramsDef.T = 0x3ee0001
    // Setting Correct N
    paramsDef.LogN = 11

    params, err := bfv.NewParametersFromLiteral(paramsDef)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }

    // optional key
    crs := []byte("eurecom")

    P0 := &Party_s{params, bfv.NewEncoder(params),nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
    P1 := &Party_s{params, bfv.NewEncoder(params),nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
    PPool := []*Party_s{P0, P1}

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

    Enrollment := &Enrollment_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk), nil, nil, nil, nil}
	BIP0 := &BIP_s{params, bfv.NewEvaluator(params, evk), nil}
	BIP1 := &BIP_s{params, bfv.NewEvaluator(params, evk), nil}
	Gate := &Gate_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk), nil, nil, nil, nil}

    // REFERENCE AND PROBE
    lfwQ := quantizeFeatures(borders, lfw)
    fmt.Println("LFWQ", lfwQ)

    refTemp := refTemplate(lfwQ, mfip)
    fmt.Println("refTemp", refTemp)

    permutations := genPermutationsConcat(SEED, NFEAT, NROWS)
    permutationsInv := getPermutationsInverse(permutations)
    fmt.Println("permutations", permutations)
    fmt.Println("permutationsInv", permutationsInv)

    quantizedProb := quantizeFeatures(borders , lfwProb)
    fmt.Println("quantizedProb", quantizedProb)

    // Do this for the BIP.c_selection
    ctxtSelection := Enrollment.encryptPermutedRefTempSingleCT(refTemp, permutations)

    BIP0.c_selection = ctxtSelection
    BIP1.c_selection = ctxtSelection

    permProbeTemp := genPermProbeTemplateFromPermInv(quantizedProb, permutationsInv, NROWS);
    permProbeTempMask := getPermutedProbeTempMask(permProbeTemp, Enrollment.params.N())
    fmt.Println("permProbeTempMask", permProbeTempMask)

    // CHANGE THIS
    Enrollment.Y_1 = mfip
    Enrollment.Y_2 = mfip

    Gate.col_selection = permProbeTempMask

    start := time.Now()
    FinalScore1 := P0.getFinalScoreCT(PPool, BIP0, Gate.col_selection)
    //FinalScore2 := P1.getFinalScoreCT(PPool, BIP1, Gate.col_selection)

    // Decrypt Final Score 1
    //encOut := CKSDecrypt(P0.params, PPool, FinalScore1)
    //ptres := bfv.NewPlaintext(params, params.MaxLevel())
	//P0.decryptor.Decrypt(encOut, ptres)

    //res := P0.encoder.DecodeIntNew(ptres)
    //fmt.Println("FinalScore1:", res)

    // Decrypt Final Score 2
    //encOut = CKSDecrypt(P1.params, PPool, FinalScore2)
    //ptres = bfv.NewPlaintext(params, params.MaxLevel())
	//P0.decryptor.Decrypt(encOut, ptres)

    //res = P1.encoder.DecodeIntNew(ptres)
    //fmt.Println("FinalScore2:", res)

    addedResult := FinalScore1 //BIP0.AddCiphertextsSingle(FinalScore1, FinalScore2)

    // Decrypt Added Score
    encOut := CKSDecrypt(P0.params, PPool, addedResult)
    ptres := bfv.NewPlaintext(params, params.MaxLevel())
	P0.decryptor.Decrypt(encOut, ptres)

    res := P0.encoder.DecodeIntNew(ptres)
    fmt.Println("addedResult:", res[:20])
    end := time.Now()
    fmt.Println("time", end.Sub(start))

    //fmt.Print("Checking if decrypted plaintexts for P0 and P1 are the same: ")
    //fmt.Println(reflect.DeepEqual(z0, z1))

    // TEST
    //r1, r2 := 1, 1 // Should be FSS randomness
    //tPrt := testProtocol(tab1selectionsFlat, tab2selectionsFlat, bmapFlat, r1, r2)
    //fmt.Println("Protocol test:", tPrt)

    //// ROW SELECTION AND MASKING
    //selections := make([]int, 0)
    //SELTIMES := 3
    //for i:=0; i<SELTIMES; i++ {
    //    selections = append(selections, rand.Intn(NROWS))
    //}
    //fmt.Printf("selections: %v\n", selections)

    //tab1selections := selectRows(Enrollment.Y_1, selections)
    //tab2selections := selectRows(Enrollment.Y_2, selections)

    //fmt.Println("selection1")
    //printMatrix(tab1selections)
    //fmt.Println("selection2")
    //printMatrix(tab2selections)

    //tab1selectionsFlat := flatten(tab1selections)
    //tab2selectionsFlat := flatten(tab2selections)
    ////tab1selectionsFlat := flatten(tab1.([][]int64))
    ////tab2selectionsFlat := flatten(tab2.([][]int64))

    //fmt.Println("tab1selectionsFlat")
    //fmt.Println(tab1selectionsFlat)
    //fmt.Println("tab2selectionsFlat")
    //fmt.Println(tab2selectionsFlat)

    //tab1selectionsFlatC := Enrollment.EncryptFlatSingle(tab1selectionsFlat)
    //encOut := CKSDecrypt(P0.params, PPool, tab1selectionsFlatC)
	//ptres := bfv.NewPlaintext(P0.params, P0.params.MaxLevel())
	//P0.decryptor.Decrypt(encOut, ptres)

	//res := P0.encoder.DecodeIntNew(ptres)
    //fmt.Println("tab1selectionsFlat Decr :", res[:50])

    //tab2selectionsFlatC := Enrollment.EncryptFlatSingle(tab2selectionsFlat)
    //encOut = CKSDecrypt(P1.params, PPool, tab2selectionsFlatC)
	//ptres = bfv.NewPlaintext(P1.params, P1.params.MaxLevel())
	//P0.decryptor.Decrypt(encOut, ptres)

	//res = P0.encoder.DecodeIntNew(ptres)
    //fmt.Println("tab2selectionsFlat Decr :", res[:50])

    //BIP0.c_selection = tab1selectionsFlatC
    //BIP1.c_selection = tab2selectionsFlatC

    //// Do this in gate and encrypt it
    //b := genRandInexes(len(selections), NROWS)
	//fmt.Println("Selected Columns", b)
    //bmapFlat := genFlatIndexMaps(b, NROWS)
    //Gate.col_selection = bmapFlat
    //fmt.Println("Flat mask:", Gate.col_selection)

    //// Testing with the unencrypted values
    //r1, r2 := 1, 1
    //tPrt := testProtocol(tab1selectionsFlat, tab2selectionsFlat, bmapFlat, r1, r2)
    //fmt.Println("Protocol test:", tPrt)

    //FinalScore1 := P0.getFinalScoreCT(PPool, BIP0, Gate.col_selection)
    //FinalScore2 := P1.getFinalScoreCT(PPool, BIP1, Gate.col_selection)

    //// Decrypt Final Score 1
    //encOut = CKSDecrypt(P0.params, PPool, FinalScore1)
    //ptres = bfv.NewPlaintext(params, params.MaxLevel())
	//P0.decryptor.Decrypt(encOut, ptres)

    //res = P0.encoder.DecodeIntNew(ptres)
    //fmt.Println("FinalScore1:", res)

    //// Decrypt Final Score 2
    ////encOut = CKSDecrypt(P1.params, PPool, FinalScore2)
    ////ptres = bfv.NewPlaintext(params, params.MaxLevel())
	////P0.decryptor.Decrypt(encOut, ptres)

    ////res = P1.encoder.DecodeIntNew(ptres)
    ////fmt.Println("FinalScore2:", res)

    //addedResult := BIP0.AddCiphertextsSingle(FinalScore1, FinalScore2)

    //// Decrypt Added Score
    //encOut = CKSDecrypt(P0.params, PPool, addedResult)
    //ptres = bfv.NewPlaintext(params, params.MaxLevel())
	//P0.decryptor.Decrypt(encOut, ptres)

    //res = P0.encoder.DecodeIntNew(ptres)
    //fmt.Println("addedResult:", res[:20])

    ////fmt.Print("Checking if decrypted plaintexts for P0 and P1 are the same: ")
    ////fmt.Println(reflect.DeepEqual(z0, z1))
}
