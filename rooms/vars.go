package rooms

import (
	"errors"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ROOM VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetVariable sets a Room variable.
func (r *Room) SetVariable(key string, value interface{}) {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return
	}

	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return
	}
	r.vars[key] = value
	r.mux.Unlock()

	//
	return
}

// SetVariables sets all the specified Room variables at once.
func (r *Room) SetVariables(values map[string]interface{}) {
	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return
	}
	for key, val := range values {
		r.vars[key] = val
	}
	r.mux.Unlock()

	//
	return
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

// GetVariables gets all the specified (or all if not) Room variables as a map[string]interface{}.
func (r *Room) GetVariables(keys []string) (map[string]interface{}, error) {
	var value map[string]interface{} = make(map[string]interface{})
	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return nil, errors.New("Room '" + r.name + "' does not exist")
	}
	if keys == nil || len(keys) == 0 {
		value = r.vars
	} else {
		for i := 0; i < len(keys); i++ {
			value[keys[i]] = r.vars[keys[i]]
		}
	}
	r.mux.Unlock()

	//
	return value, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetVariable sets a RoomUser's variable.
func (r *RoomUser) SetVariable(key string, value interface{}) {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return
	}
	(*r.mux).Lock()
	(*r.vars)[key] = value
	(*r.mux).Unlock()
}

// SetVariables sets all the specified RoomUser's variables at once.
func (r *RoomUser) SetVariables(values map[string]interface{}) {
	(*r.mux).Lock()
	for key, val := range values {
		(*r.vars)[key] = val
	}
	(*r.mux).Unlock()
}

// GetVariable gets a RoomUser's variable.
func (r *RoomUser) GetVariable(key string) interface{} {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return errors.New("*Room.GetUserVariable() requires a variable name")
	}

	var value interface{}

	(*r.mux).Lock()
	value = (*r.vars)[key]
	(*r.mux).Unlock()

	//
	return value
}

// GetVariables gets all the specified (or all if not) RoomUser's variables as a map[string]interface{}.
func (r *RoomUser) GetVariables(keys []string) map[string]interface{} {
	var value map[string]interface{} = make(map[string]interface{})

	if keys == nil || len(keys) == 0 {
		(*r.mux).Lock()
		value = *r.vars
		(*r.mux).Unlock()
	}else{
		(*r.mux).Lock()
		for i := 0; i < len(keys); i++ {
			value[keys[i]] = (*r.vars)[keys[i]]
		}
		(*r.mux).Unlock()
	}

	//
	return value
}
