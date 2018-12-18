package users

import (
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
	//MAKE CLIENT MESSAGE
	resp := make(map[string]interface{})
	resp["k"] = key
	resp["v"] = value

	u.mux.Lock()
	if _, ok := u.conns[connID]; !ok {
		u.mux.Unlock()
		return
	}
	(*u.conns[connID]).vars[key] = value
	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionSetVariable, resp, helpers.NewError("", 0))
	(*u.conns[connID]).socket.WriteJSON(clientResp)
	u.mux.Unlock()
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
	u.mux.Lock()
	if _, ok := u.conns[connID]; !ok {
		u.mux.Unlock()
		return
	}
	for key, val := range values {
		(*u.conns[connID]).vars[key] = val
	}
	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionSetVariables, values, helpers.NewError("", 0))
	(*u.conns[connID]).socket.WriteJSON(clientResp)
	u.mux.Unlock()
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
