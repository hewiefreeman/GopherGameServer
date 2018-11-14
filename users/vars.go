package users

import (
	"errors"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetVariable sets a User variable.
func (u *User) SetVariable(key string, value interface{}) {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return
	}
	u.mux.Lock()
	u.vars[key] = value
	u.mux.Unlock()

	//
	return nil
}

// SetVariables sets all the specified User variables at once.
func (u *User) SetVariables(values map[string]interface{}) {
	u.mux.Lock()
	for key, val := range values{
		u.vars[key] = val
	}
	u.mux.Unlock()

	//
	return nil
}

// GetVariable gets one of the User's variables by it's key.
func (u *User) GetVariable(key string) interface{} {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return nil
	}
	u.mux.Lock()
	theVar := u.vars[key]
	u.mux.Unlock()

	//
	return theVar
}

// GetVariable gets one of the User's variables by it's key.
func (u *User) GetVariables(keys []string) map[string]interface{} {
	var value map[string]interface{} = make(map[string]interface{})
	if(keys == nil || len(keys) == 0) {
		u.mux.Lock()
		value = u.vars
		u.mux.Unlock()
	}else{
		u.mux.Lock()
		for i := 0; i < len(keys); i++ {
			value[keys[i]] = u.vars[keys[i]]
		}
		u.mux.Unlock()
	}

	//
	return value
}
