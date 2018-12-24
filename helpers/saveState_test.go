package helpers

import (
	"fmt"
	"testing"
)

func TestSaveState(*testing.T) {
	e := make(map[string]interface{})
	e["wtf"] = make(map[string]interface{})
	e["wtf"].(map[string]interface{})["innerWtf"] = "howdy do!"
	err := saveState(e)
	if err != nil {
		fmt.Println(err)
	}
}
