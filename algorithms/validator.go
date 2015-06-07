package algorithms

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func Validate(f string) {
	csvfile, err := os.Open(f)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = 2
	records, err := reader.ReadAll()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// sanity check, display to standard output
	var timeseries []float64 = make([]float64, 0)

	for _, each := range records {
		var input, _ = strconv.Atoi(each[0])
		var expected, _ = strconv.Atoi(each[1])

		timeseries = append(timeseries, float64(input))
		var calculated = AutoAlg(timeseries, 3)

		var diff = (float64(expected) - float64(calculated)) / (float64(expected))

		if diff < -1 || diff > 1 {
			fmt.Printf("Got %d, expected %d and calculated %d (%v)\n", input, expected, calculated, diff)
		}

	}
}
