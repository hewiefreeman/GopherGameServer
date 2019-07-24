package core

import (
	"errors"
	"github.com/hewiefreeman/GopherGameServer/helpers"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetVariable sets a User variable. The client API of the User will also receive these changes. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to the inviting User. This must
// be provided when setting a User's variables with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) SetVariable(key string, value interface{}, connID string) {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return
	} else if multiConnect && len(connID) == 0 {
		return
	} else if !multiConnect {
		connID = "1"
	}

	// Set the variable
	u.mux.Lock()
	if _, ok := u.conns[connID]; !ok {
		u.mux.Unlock()
		return
	}
	(*u.conns[connID]).vars[key] = value
	socket := (*u.conns[connID]).socket
	u.mux.Unlock()

	//MAKE CLIENT MESSAGE
	resp := map[string]interface{}{
		"k": key,
		"v": value,
	}
	clientResp := helpers.MakeClientResponse(helpers.ClientActionSetVariable, resp, helpers.NoError())

	//SEND RESPONSE TO CLIENT
	socket.WriteJSON(clientResp)
}

// SetVariables sets all the specified User variables at once. The client API of the User will also receive these changes. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to the inviting User. This must
// be provided when setting a User's variables with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) SetVariables(values map[string]interface{}, connID string) {
	//REJECT INCORRECT INPUT
	if values == nil || len(values) == 0 {
		return
	} else if multiConnect && len(connID) == 0 {
		return
	} else if !multiConnect {
		connID = "1"
	}

	// Set the variables
	u.mux.Lock()
	if _, ok := u.conns[connID]; !ok {
		u.mux.Unlock()
		return
	}
	for key, val := range values {
		(*u.conns[connID]).vars[key] = val
	}
	socket := (*u.conns[connID]).socket
	u.mux.Unlock()

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionSetVariables, values, helpers.NoError())
	socket.WriteJSON(clientResp)

}

// GetVariable gets one of the User's variables by it's key. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to the inviting User. This must
// be provided when getting a User's variables with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) GetVariable(key string, connID string) interface{} {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return nil
	}

	u.mux.Lock()
	val := (*u.conns[connID]).vars[key]
	u.mux.Unlock()

	//
	return val
}

// GetVariables gets the specified (or all if nil) User variables as a map[string]interface{}.  If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to the inviting User. This must
// be provided when getting a User's variables with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) GetVariables(keys []string, connID string) map[string]interface{} {
	var value map[string]interface{} = make(map[string]interface{})
	if keys == nil || len(keys) == 0 {
		u.mux.Lock()
		value = (*u.conns[connID]).vars
		u.mux.Unlock()
	} else {
		u.mux.Lock()
		for i := 0; i < len(keys); i++ {
			value[keys[i]] = (*u.conns[connID]).vars[keys[i]]
		}
		u.mux.Unlock()
	}

	//
	return value
}

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
