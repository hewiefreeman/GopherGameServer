package users

import (
	"github.com/hewiefreeman/GopherGameServer/helpers"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND A PRIVATE MESSAGE   ////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// PrivateMessage sends a private message to another User by name.
func (u *User) PrivateMessage(userName string, message string) error {
	user, userErr := Get(userName)
	if userErr != nil {
		return userErr
	}

	//CONSTRUCT MESSAGE
	theMessage := make(map[string]interface{})
	theMessage[helpers.ServerActionPrivateMessage] = make(map[string]interface{})
	theMessage[helpers.ServerActionPrivateMessage].(map[string]interface{})["a"] = u.name
	theMessage[helpers.ServerActionPrivateMessage].(map[string]interface{})["m"] = message

	sendErr := user.socket.WriteJSON(theMessage)
	if sendErr != nil {
		return sendErr
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND A DATA MESSAGE   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// DataMessage sends a data message directly to the User.
func (u *User) DataMessage(data interface{}) error {
	//CONSTRUCT MESSAGE
	message := make(map[string]interface{})
	message[helpers.ServerActionDataMessage] = data

	//SEND MESSAGE TO USERS
	u.socket.WriteJSON(message)

	//
	return nil
}
