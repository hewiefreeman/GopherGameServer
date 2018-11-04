package users

import (
	"github.com/hewiefreeman/GopherGameServer/helpers"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND A PRIVATE MESSAGE   ////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Sends a private message to another User by name.
func (u *User) PrivateMessage(userName string, message string) error {
	user, userErr := Get(userName);
	if(userErr != nil){ return userErr }

	//CONSTRUCT MESSAGE
	theMessage := make(map[string]interface{});
	theMessage[helpers.ServerActionPrivateMessage] = make(map[string]interface{});
	theMessage[helpers.ServerActionPrivateMessage].(map[string]interface{})["a"] = u.name;
	theMessage[helpers.ServerActionPrivateMessage].(map[string]interface{})["m"] = message;

	sendErr := user.socket.WriteJSON(theMessage);
	if(sendErr != nil){ return sendErr; }

	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND A DATA MESSAGE   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func (u *User) DataMessage(data interface{}) error {
	//CONSTRUCT MESSAGE
	message := make(map[string]interface{});
	message[helpers.ServerActionDataMessage] = data;

	//SEND MESSAGE TO USERS
	u.socket.WriteJSON(message);

	//
	return nil;
}
