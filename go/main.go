package main

import (
	"fmt"
	"math/rand"
	//"reflect"

	"github.com/tuneinsight/lattigo/v4/bfv"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/utils"
)

func main() {
    //READING THE DATA AND TABLE CONVERSION
    tab1path := "./lookupTables/Rand/Rand_nB_3_dimF_512.csv"
    tab2path := "./lookupTables/MFIP-Rand/MFIPSubRand_nB_3_dimF_512.csv"
    tab1, err := readCSVToArray(tab1path)
    tab2, err := readCSVToArray(tab2path)
    if err != nil {
        fmt.Println("Error reading one of the csv files:", err)
    }
    fmt.Println("tab1")
    printMatrix(tab1.([][]int64))
    fmt.Println("tab2")
    printMatrix(tab2.([][]int64))

    size := len(tab1.([][]int64))

    //MULTI PARTY MULTI BIP HE
    paramsDef := bfv.PN13QP218
    // Set the propper T value instead of a default later
    paramsDef.T = 0x3ee0001
    // Setting Correct N
    paramsDef.LogN = 13

    params, err := bfv.NewParametersFromLiteral(paramsDef)
    if err != nil {
        fmt.Println(fmt.Errorf(err.Error()))
    }

    // optional key
    crs := []byte("eurecom")

    P0 := &Party_s{params, bfv.NewEncoder(params),nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
    P1 := &Party_s{params, bfv.NewEncoder(params),nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
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
	evk := rlwe.EvaluationKey{Rlk: rlk, Rtks: nil}

    Enrollment := &Enrollment_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk), nil, nil, nil, nil}
	BIP0 := &BIP_s{params, bfv.NewEvaluator(params, evk), nil, nil , nil}
	BIP1 := &BIP_s{params, bfv.NewEvaluator(params, evk), nil, nil , nil}
	Gate := &Gate_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk), nil, nil, nil, nil}


    Enrollment.Y_1 = tab1.([][]int64)
    Enrollment.Y_2 = tab2.([][]int64)

    // ROW SELECTION AND MASKING
    selections := []int{rand.Intn(size), rand.Intn(size), rand.Intn(size)}
    fmt.Printf("selections: %v\n", selections)

    tab1selections := selectRows(Enrollment.Y_1, selections)
    tab2selections := selectRows(Enrollment.Y_2, selections)

    fmt.Println("selection1")
    printMatrix(tab1selections)
    fmt.Println("selection2")
    printMatrix(tab2selections)

    tab1selectionsFlat := flatten(tab1selections)
    tab2selectionsFlat := flatten(tab2selections)
    fmt.Println("tab1selectionsFlat")
    fmt.Println(tab1selectionsFlat)
    fmt.Println("tab2selectionsFlat")
    fmt.Println(tab2selectionsFlat)

    //tab1selectionsFlatC := Enrollment.EncryptFlatSingle(tab1selectionsFlat)
    //tab2selectionsFlatC := Enrollment.EncryptFlatSingle(tab2selectionsFlat)
    ptxt1 := Enrollment.encoder.EncodeNew(tab1selectionsFlat, params.MaxLevel())
    tab1selectionsFlatC := Enrollment.encryptor.EncryptNew(ptxt1)

    ptxt2 := Enrollment.encoder.EncodeNew(tab1selectionsFlat, params.MaxLevel())
    tab2selectionsFlatC := Enrollment.encryptor.EncryptNew(ptxt2)

    //fmt.Println("tab1selectionsFlatC")
    //fmt.Println(tab1selectionsFlatC)
    //fmt.Println("tab2selectionsFlatC")
    //fmt.Println(tab2selectionsFlatC)
    finalScorePT := P0.decryptor.DecryptNew(tab2selectionsFlatC)
    finalScoreVec := Enrollment.encoder.DecodeUintNew(finalScorePT)
    fmt.Println("decoded ta1FC", finalScoreVec[:50])

    finalScorePT = P0.decryptor.DecryptNew(tab2selectionsFlatC)
    finalScoreVec = Enrollment.encoder.DecodeUintNew(finalScorePT)
    fmt.Println("decoded ta2FC", finalScoreVec[:50])

    BIP0.c_selection = tab1selectionsFlatC
    BIP1.c_selection = tab2selectionsFlatC

    // Do this in gate and encrypt it
    b := genRandInexes(len(selections), size)
	fmt.Println("Selected Columns", b)
    bmapFlat := genFlatIndexMaps(b, size)
    Gate.col_selection = bmapFlat
    fmt.Println("Flat mask:", Gate.col_selection)

    FinalScore1 := P0.getFinalScoreCT(BIP0, Gate.col_selection)
    FinalScore2 := P1.getFinalScoreCT(BIP1, Gate.col_selection)

    fmt.Println("FinalScore1", FinalScore1)
    fmt.Println("FinalScore2", FinalScore2)

    // Does bip do this or should the parties do this
    //result1 := BIP0.compCFlatRowsTimesMasks(BIP0.c_selection, Gate.col_selection)
    //result2 := BIP1.compCFlatRowsTimesMasks(BIP1.c_selection, Gate.col_selection)

    //fmt.Println("result1")
    //fmt.Println(result1)
    //fmt.Println("result2")
    //fmt.Println(result2)

    //// Add The ciphertexts
    //addedResult := BIP0.AddCiphertexts(result1, result2)
    //fmt.Println("Added Result")
    //fmt.Println(addedResult)

    //P0.c_z = addedResult
    //P1.c_z = addedResult

    //P0.c1sShares = P0.C1ShareDecrypt(P0.c_z)
    //P1.c1sShares = P1.C1ShareDecrypt(P1.c_z)

    //fmt.Println("c1sSharesP0")
    //fmt.Println(P0.c1sShares)
    //fmt.Println("c1sSharesP1")
    //fmt.Println(P1.c1sShares)

    //z_0 := P0.AggregateAndDecrypt(P1.c1sShares)
    //z_1 := P1.AggregateAndDecrypt(P0.c1sShares)

    //fmt.Print("Checking if decrypted plaintexts for P0 and P1 are the same: ")
    //fmt.Println(reflect.DeepEqual(z_0, z_1))
}
