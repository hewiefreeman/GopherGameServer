package rooms

import (
	"encoding/json"
	"errors"
)

// These represent the types of room messages the server sends.
const (
	MessageTypeChat = iota
	MessageTypeServer
)

// These are the sub-types that a MessageTypeServer will come with. Ordered by their visible priority for your UI.
const (
	ServerMessageGame = iota
	ServerMessageNotice
	ServerMessageImportant
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHAT MESSAGES   //////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Sends a chat message to the Room.
func (r *Room) ChatMessage(author string, message interface{}) error {
	//REJECT INCORRECT INPUT
	if(len(author) == 0){
		return errors.New("*Room.ChatMessage() requires an author")
	}else if(message == nil){
		return errors.New("*Room.ChatMessage() requires a message")
	}

	return r.sendMessage(MessageTypeChat, 0, nil, author, message);
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER MESSAGES   ////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Sends a server message to the Room.
func (r *Room) ServerMessage(message interface{}, messageType int, recipients []string) error {
	if(message == nil){ return errors.New("*Room.ServerMessage() requires a message") }

	return r.sendMessage(MessageTypeServer, messageType, recipients, "", message);
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   DATA MESSAGES   //////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Sends a data message to the Room.
func (r *Room) DataMessage(message interface{}, recipients []string) error {
	//GET USER MAP
	userMap, err := r.GetUserMap();
	if(err != nil){ return err; }

	//CONSTRUCT MESSAGE
	theMessage := make(map[string]interface{});
	theMessage["d"] = message; // Data messages are labeled "d"

	//MARSHAL THE MESSAGE
	jsonStr, marshErr := json.Marshal(theMessage);
	if(marshErr != nil){ return marshErr; }

	//SEND MESSAGE TO USERS
	if(recipients == nil || len(recipients) == 0){
		for _, v := range userMap { v.socket.WriteJSON(jsonStr); }
	}else{
		for i := 0; i < len(recipients); i++ {
			if u, ok := userMap[recipients[i]]; ok { u.socket.WriteJSON(jsonStr); }
		}
	}

	//
	return nil;
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   SENDING MESSAGES   ///////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Room) sendMessage(mt int, st int, rec []string, a string, m interface{}) error {
	//GET USER MAP
	userMap, err := r.GetUserMap();
	if(err != nil){ return err; }

	//CONSTRUCT MESSAGE
	message := make(map[string]interface{});
	message["m"] = make(map[string]interface{}); // Room messages are labeled "m"
	if(mt == MessageTypeServer){ message["m"].(map[string]interface{})["s"] = st; } // Server messages come with a sub-type
	if(len(a) > 0 && mt != MessageTypeServer){ message["m"].(map[string]interface{})["a"] = a; } // Non-server messages have authors
	message["m"].(map[string]interface{})["m"] = m; // The message

	//MARSHAL THE MESSAGE
	jsonStr, marshErr := json.Marshal(message);
	if(marshErr != nil){ return marshErr; }

	//SEND MESSAGE TO USERS
	if(rec == nil || len(rec) == 0){
		for _, v := range userMap { v.socket.WriteJSON(jsonStr); }
	}else{
		for i := 0; i < len(rec); i++ {
			if u, ok := userMap[rec[i]]; ok { u.socket.WriteJSON(jsonStr); }
		}
	}

	return nil;
}
