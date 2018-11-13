package rooms

import (
	"errors"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ROOM VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetVariable sets a Room variable.
func (r *Room) SetVariable(key string, value interface{}) error {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return errors.New("*Room.SetVariable() requires a key")
	}

	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return errors.New("Room '" + r.name + "' does not exist")
	}
	r.vars[key] = value
	r.mux.Unlock()

	//
	return nil
}

// GetVariable gets one of the Room's variables.
func (r *Room) GetVariable(key string) (interface{}, error) {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return nil, errors.New("*Room.GetVariable() requires a key")
	}

	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return nil, errors.New("Room '" + r.name + "' does not exist")
	}
	value := r.vars[key]
	r.mux.Unlock()

	//
	return value, nil
}

// GetVariables gets all the Room variables as a map[string]interface{}.
func (r *Room) GetVariables() (map[string]interface{}, error) {
	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return nil, errors.New("Room '" + r.name + "' does not exist")
	}
	value := r.vars
	r.mux.Unlock()

	//
	return value, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetUserVariable sets a User's variable using their name.
func (r *RoomUser) SetVariable(varName string, value interface{}) {
	//REJECT INCORRECT INPUT
	if len(varName) == 0 {
		return
	}
	(*r.mux).Lock()
	(*r.vars)[varName] = value
	(*r.mux).Unlock()
}

// GetUserVariable gets a User's variable.
func (r *RoomUser) GetVariable(varName string) interface{} {
	//REJECT INCORRECT INPUT
	if len(varName) == 0 {
		return errors.New("*Room.GetUserVariable() requires a variable name")
	}

	var value interface{}

	(*r.mux).Lock()
	value = (*r.vars)[varName]
	(*r.mux).Unlock()

	//
	return value
}

// GetUserVariables gets all the User's variables using their name.
func (r *RoomUser) GetVariables() map[string]interface{} {
	var value map[string]interface{}

	(*r.mux).Lock()
	value = *r.vars
	(*r.mux).Unlock()

	//
	return value
}
