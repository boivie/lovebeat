package algorithms

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func Validate() {
	scanner := bufio.NewScanner(os.Stdin)
	var timeseries []int64

	for scanner.Scan() {
		input, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			return
		}
		timeseries = append(timeseries, int64(input))
		var calculated = AutoAlg(timeseries)
		fmt.Printf("%d\n", calculated)
	}
}
