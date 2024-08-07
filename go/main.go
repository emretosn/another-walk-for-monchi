package main

import (
	"fmt"
	"math/rand"
	"reflect"

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
	BIP1 := &BIP_s{params, bfv.NewEvaluator(params, evk), nil, nil, nil}
	BIP2 := &BIP_s{params, bfv.NewEvaluator(params, evk), nil, nil, nil}
	//Gate := &Gate_s{params, bfv.NewEncoder(params), bfv.NewEncryptor(params, pk), nil, nil, nil}

    Enrollment.Y_1 = tab1.([][]int64)
    Enrollment.Y_2 = tab2.([][]int64)
	Enrollment.c_Y_1 = Enrollment.EncryptInput(Enrollment.Y_1)
	Enrollment.c_Y_2 = Enrollment.EncryptInput(Enrollment.Y_2)
    BIP1.c_Y = Enrollment.c_Y_1
    BIP2.c_Y = Enrollment.c_Y_2

    fmt.Println("tab1 encrypted")
    printMatrix(Enrollment.c_Y_1)

    fmt.Println("tab2 encrypted")
    printMatrix(Enrollment.c_Y_2)

    // ROW SELECTION AND MASKING
    selections := []int{rand.Intn(size), rand.Intn(size), rand.Intn(size)}
    fmt.Printf("selections: %v\n", selections)
    tab1selections := selectRows(Enrollment.c_Y_1, selections)
    tab2selections := selectRows(Enrollment.c_Y_2, selections)

    fmt.Println("selection1")
    printMatrix(tab1selections)
    fmt.Println("selection2")
    printMatrix(tab2selections)


    b := genRandInexes(len(selections), size)
	fmt.Println("Selected Columns", b)

    bmap := genIndexMaps(b, size)

    tab1selectionsFlat := flatten(tab1selections)
    tab2selectionsFlat := flatten(tab2selections)
    bmapFlat := flatten(bmap)
    fmt.Println("Flat mask:", bmapFlat)

    result1 := compFlatRowsTimesMasks(tab1selectionsFlat, bmapFlat)
    result2 := compFlatRowsTimesMasks(tab2selectionsFlat, bmapFlat)

    fmt.Println("result1")
    fmt.Println(result1)
    fmt.Println("result2")
    fmt.Println(result2)

    addedResult := BIP1.AddCiphertexts(result1, result2)
    fmt.Println("Added Result")
    fmt.Println(addedResult)

    //fmt.Println("Added coefs")
    //for _, res := range addedResult {
    //    if res != nil {
    //        fmt.Printf("%v ", res.Value[0])
    //    } else {
    //        fmt.Printf("%v ", nil)
    //    }
    //}


    P0.c_z = addedResult
    P1.c_z = addedResult

    P0.c1sShares = P0.C1ShareDecrypt(P0.c_z)
    P1.c1sShares = P1.C1ShareDecrypt(P1.c_z)

    fmt.Println("c1sSharesP0")
    fmt.Println(P0.c1sShares)
    fmt.Println("c1sSharesP1")
    fmt.Println(P1.c1sShares)

    z_0 := P0.AggregateAndDecrypt(P1.c1sShares)
    z_1 := P1.AggregateAndDecrypt(P0.c1sShares)

    fmt.Print("Checking if decrypted plaintexts for P0 and P1 are the same: ")
    fmt.Println(reflect.DeepEqual(z_0, z_1))
}
