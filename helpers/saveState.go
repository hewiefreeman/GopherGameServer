package helpers

import (
	"io/ioutil"
	"encoding/json"
)

func saveState(stateObj map[string]interface{}, saveFolder string) error {
	//WRITE THE STATE
	stateStr, err := json.Marshal(stateObj)
	if err != nil {
		return err
	}
	state := []byte(stateStr)
	err = ioutil.WriteFile(saveFolder, state, 0644)
	if err != nil {
		return err
	}

	//
	return nil
}
