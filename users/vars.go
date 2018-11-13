package users

import (
	"errors"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetVariable sets a User variable.
func (u *User) SetVariable(key string, value interface{}) error {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return errors.New("*User.SetVariable() requires a key")
	}

	u.mux.Lock()
	u.vars[key] = value
	u.mux.Unlock()

	//
	return nil
}

// GetVariable gets one of the User's variables.
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
