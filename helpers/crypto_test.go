package helpers

import (
	"fmt"
	"testing"
)

/*func TestHashNumber(t *testing.T) {

	// The amount of databases
	amount := 5
	lst := make([]int, amount, amount)

	// Maximum ID
	max := 4294967295

	// Max/Min generated IDs
	fullMin := 0
	fullMax := 0

	for i := 0; i < 4000000; i++ {
		str, err := GenerateSecureString(32)
		if err != nil {
			fmt.Println(err)
			return
		}

		numb := UserHashNumber(str)
		if fullMin == 0 || numb < fullMin {
			fullMin = numb
		}
		if numb > fullMax {
			fullMax = numb
		}
		for i := 0; i < amount; i++ {
			if numb <= (max/amount)*(i+1) {
				lst[i]++
				break
			}
		}
	}

	for i := 0; i < len(lst); i++ {
		fmt.Println("Database", i, "would have", lst[i], "entries")
	}

	fmt.Println("Min:", fullMin)
	fmt.Println("Max:", fullMax)
}*/
