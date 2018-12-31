package users

import (
	"github.com/hewiefreeman/GopherGameServer/helpers"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND A PRIVATE MESSAGE   ////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// PrivateMessage sends a private message to another User by name.
func (u *User) PrivateMessage(userName string, message interface{}) {
	user, userErr := Get(userName)
	if userErr != nil {
		return
	}

	//CONSTRUCT MESSAGE
	theMessage := make(map[string]interface{})
	theMessage[helpers.ServerActionPrivateMessage] = make(map[string]interface{})
	theMessage[helpers.ServerActionPrivateMessage].(map[string]interface{})["f"] = u.name    // from
	theMessage[helpers.ServerActionPrivateMessage].(map[string]interface{})["t"] = user.name // to
	theMessage[helpers.ServerActionPrivateMessage].(map[string]interface{})["m"] = message   // message

	//SEND MESSAGES
	user.mux.Lock()
	for _, conn := range user.conns {
		(*conn).socket.WriteJSON(theMessage)
	}
	user.mux.Unlock()
	u.mux.Lock()
	for _, conn := range u.conns {
		(*conn).socket.WriteJSON(theMessage)
	}
	u.mux.Unlock()

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND A DATA MESSAGE   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// DataMessage sends a data message directly to the User.
func (u *User) DataMessage(data interface{}, connID string) {
	//CONSTRUCT MESSAGE
	message := make(map[string]interface{})
	message[helpers.ServerActionDataMessage] = data

	//SEND MESSAGE TO USER
	u.mux.Lock()
	if connID == "" {
		for _, conn := range u.conns {
			(*conn).socket.WriteJSON(message)
		}
	} else {
		if conn, ok := u.conns[connID]; ok {
			(*conn).socket.WriteJSON(message)
		}
	}
	u.mux.Unlock()
}
