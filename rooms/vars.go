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

// SetVariable sets a RoomUser's variable. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when manipulating a RoomUser's variables with MultiConnect enabled. Otherwise, an empty string can be used.
func (r *RoomUser) SetVariable(key string, value interface{}, connID string) {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return
	} else if multiConnect && len(connID) == 0 {
		return
	} else if !multiConnect {
		connID = "1"
	}
	(*r.mux).Lock()
	if conn, ok := r.conns[connID]; ok {
		(*(*conn).vars)[key] = value
	}
	(*r.mux).Unlock()
}

// SetVariables sets all the specified RoomUser's variables at once. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when manipulating a RoomUser's variables with MultiConnect enabled. Otherwise, an empty string can be used.
func (r *RoomUser) SetVariables(values map[string]interface{}, connID string) {
	if multiConnect && len(connID) == 0 {
		return
	} else if !multiConnect {
		connID = "1"
	}
	(*r.mux).Lock()
	if conn, ok := r.conns[connID]; ok {
		for key, val := range values {
			(*(*conn).vars)[key] = val
		}
	}
	(*r.mux).Unlock()
}

// GetVariable gets a RoomUser's variable. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when manipulating a RoomUser's variables with MultiConnect enabled. Otherwise, an empty string can be used.
func (r *RoomUser) GetVariable(key string, connID string) interface{} {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return nil
	} else if multiConnect && len(connID) == 0 {
		return nil
	} else if !multiConnect {
		connID = "1"
	}

	var value interface{}

	(*r.mux).Lock()
	if conn, ok := r.conns[connID]; ok {
		value = (*(*conn).vars)[key]
	}
	(*r.mux).Unlock()

	//
	return value
}

// GetVariables gets all the specified (or all if not) RoomUser's variables as a map[string]interface{}. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when manipulating a RoomUser's variables with MultiConnect enabled. Otherwise, an empty string can be used.
func (r *RoomUser) GetVariables(keys []string, connID string) map[string]interface{} {
	if multiConnect && len(connID) == 0 {
		return nil
	} else if !multiConnect {
		connID = "1"
	}

	var value map[string]interface{} = make(map[string]interface{})

	if keys == nil || len(keys) == 0 {
		(*r.mux).Lock()
		value = *((*r.conns[connID]).vars)
		(*r.mux).Unlock()
	} else {
		(*r.mux).Lock()
		for i := 0; i < len(keys); i++ {
			if conn, ok := r.conns[connID]; ok {
				value[keys[i]] = (*(*conn).vars)[keys[i]]
			} else {
				value[keys[i]] = nil
			}
		}
		(*r.mux).Unlock()
	}

	//
	return value
}
