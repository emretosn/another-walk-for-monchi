package main

import (
    "os"
    "fmt"
    "log"
    "strconv"
	"encoding/csv"
    "path/filepath"
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
    writer.Flush()
}

