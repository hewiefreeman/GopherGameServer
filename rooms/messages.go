package rooms

import (
	"errors"
	"github.com/hewiefreeman/GopherGameServer/helpers"
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

// ChatMessage sends a chat message to all Users in the Room.
func (r *Room) ChatMessage(author string, message interface{}) error {
	//REJECT INCORRECT INPUT
	if len(author) == 0 {
		return errors.New("*Room.ChatMessage() requires an author")
	} else if message == nil {
		return errors.New("*Room.ChatMessage() requires a message")
	}

	return r.sendMessage(MessageTypeChat, 0, nil, author, message)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER MESSAGES   ////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// ServerMessage sends a server message to the specified recipients in the Room. The parameter recipients can be nil or an empty slice
// of string. In which case, the server message will be sent to all Users in the Room.
func (r *Room) ServerMessage(message interface{}, messageType int, recipients []string) error {
	if message == nil {
		return errors.New("*Room.ServerMessage() requires a message")
	}

	return r.sendMessage(MessageTypeServer, messageType, recipients, "", message)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   DATA MESSAGES   //////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// DataMessage sends a data message to the specified recipients in the Room. The parameter recipients can be nil or an empty slice
// of string. In which case, the data message will be sent to all Users in the Room.
func (r *Room) DataMessage(message interface{}, recipients []string) error {
	//GET USER MAP
	userMap, err := r.GetUserMap()
	if err != nil {
		return err
	}

	//CONSTRUCT MESSAGE
	theMessage := make(map[string]interface{})
	theMessage[helpers.ServerActionDataMessage] = message

	//SEND MESSAGE TO USERS
	if recipients == nil || len(recipients) == 0 {
		for _, u := range userMap {
			u.mux.Lock()
			for _, conn := range u.conns {
				(*conn).socket.WriteJSON(theMessage)
			}
			u.mux.Unlock()
		}
	} else {
		for i := 0; i < len(recipients); i++ {
			if u, ok := userMap[recipients[i]]; ok {
				u.mux.Lock()
				for _, conn := range u.conns {
					(*conn).socket.WriteJSON(theMessage)
				}
				u.mux.Unlock()
			}
		}
	}

	//
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   SENDING MESSAGES   ///////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Room) sendMessage(mt int, st int, rec []string, a string, m interface{}) error {
	//GET USER MAP
	userMap, err := r.GetUserMap()
	if err != nil {
		return err
	}

	//CONSTRUCT MESSAGE
	message := make(map[string]interface{})
	message[helpers.ServerActionRoomMessage] = make(map[string]interface{})
	if mt == MessageTypeServer {
		message[helpers.ServerActionRoomMessage].(map[string]interface{})["s"] = st
	} // Server messages come with a sub-type
	if len(a) > 0 && mt != MessageTypeServer {
		message[helpers.ServerActionRoomMessage].(map[string]interface{})["a"] = a
	} // Non-server messages have authors
	message[helpers.ServerActionRoomMessage].(map[string]interface{})["m"] = m // The message

	//SEND MESSAGE TO USERS
	if rec == nil || len(rec) == 0 {
		for _, u := range userMap {
			u.mux.Lock()
			for _, conn := range u.conns {
				(*conn).socket.WriteJSON(message)
			}
			u.mux.Unlock()
		}
	} else {
		for i := 0; i < len(rec); i++ {
			if u, ok := userMap[rec[i]]; ok {
				u.mux.Lock()
				for _, conn := range u.conns {
					(*conn).socket.WriteJSON(message)
				}
				u.mux.Unlock()
			}
		}
	}

	return nil
}
