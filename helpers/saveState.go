package helpers

import (
	"encoding/json"
	"io/ioutil"
)

func getState() map[string]interface{} {
	//GET THE SERVER'S STATE
	state := make(map[string]interface{})

	//
	return state
}

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
