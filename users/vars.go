package users

import (
	"github.com/hewiefreeman/GopherGameServer/helpers"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetVariable sets a User variable. The client API of the User will also recieve these changes.
func (u *User) SetVariable(key string, value interface{}) {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return
	}
	u.mux.Lock()
	u.vars[key] = value
	u.mux.Unlock()

	//MAKE CLIENT MESSAGE
	resp := make(map[string]interface{})
	resp["k"] = key
	resp["v"] = value

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionSetVariable, resp, nil)
	u.socket.WriteJSON(clientResp)
}

// SetVariables sets all the specified User variables at once. The client API of the User will also recieve these changes.
func (u *User) SetVariables(values map[string]interface{}) {
	//REJECT INCORRECT INPUT
	if values == nil || len(values) == 0 {
		return
	}
	u.mux.Lock()
	for key, val := range values{
		u.vars[key] = val
	}
	u.mux.Unlock()

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionSetVariables, values, nil)
	u.socket.WriteJSON(clientResp)
}

// GetVariable gets one of the User's variables by it's key.
func (u *User) GetVariable(key string) interface{} {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return nil
	}
	u.mux.Lock()
	val := u.vars[key]
	u.mux.Unlock()

	//
	return val
}

// GetVariables gets the specified (or all if nil) User variables as a map[string]interface{}.
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
