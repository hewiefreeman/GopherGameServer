package helpers

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// SaveState is only for internal Gopher Server mechanics.
func SaveState(stateObj map[string]interface{}, saveFolder string) error {
	//WRITE THE STATE
	stateStr, err := json.Marshal(stateObj)
	if err != nil {
		return err
	}
	state := []byte(stateStr)
	err = ioutil.WriteFile(saveFolder+"/Gopher Recovery - "+time.Now().Format("2006-01-02 15-04-05")+".grf", state, 0644)
	if err != nil {
		return err
	}

	//
	return nil
}
