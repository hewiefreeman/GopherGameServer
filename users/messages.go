package users

import (
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"encoding/json"
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
	theMessage["p"] = make(map[string]interface{});
	theMessage["p"].(map[string]interface{})["a"] = u.name;
	theMessage["p"].(map[string]interface{})["m"] = message;

	//MARSHAL MESSAGE INTO JSON
	jsonStr, marshErr := json.Marshal(theMessage);
	if(marshErr != nil){ return marshErr; }

	sendErr := user.socket.WriteJSON(jsonStr);
	if(sendErr != nil){ return sendErr; }

	return nil;
}
