package helpers

import (
	"testing"
	"fmt"
)

func TestHashNumber(t *testing.T) {

	amount := 40
	lst := make([]int, amount, amount)

	for i := 0; i < 1000000; i++ {
		str, err := GenerateSecureString(3)
		if err != nil {
			fmt.Println(err)
			return
		}

		numb := HashNumber(str, amount)
		lst[numb]++
	}

	for i := 0; i < len(lst); i++ {
		fmt.Println("Database", i, "would have", lst[i], "entries")
	}
}
