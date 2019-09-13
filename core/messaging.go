package core

import (
	"errors"
	"github.com/gorilla/websocket"
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

var (
	privateMessageCallback    func(*User, *User, interface{})
	privateMessageCallbackSet bool
	chatMessageCallback       func(string, *Room, interface{})
	chatMessageCallbackSet    bool
	serverMessageCallback     func(*Room, int, interface{})
	serverMessageCallbackSet  bool
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Messaging Users   ///////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// PrivateMessage sends a private message to another User by name.
func (u *User) PrivateMessage(userName string, message interface{}) {
	user, userErr := GetUser(userName)
	if userErr != nil {
		return
	}

	//CONSTRUCT MESSAGE
	theMessage := map[string]map[string]interface{}{
		helpers.ServerActionPrivateMessage: {
			"f": u.name,    // from
			"t": user.name, // to
			"m": message,
		},
	}

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

	if privateMessageCallbackSet {
		privateMessageCallback(u, user, message)
	}

	return
}

// DataMessage sends a data message directly to the User.
func (u *User) DataMessage(data interface{}, connID string) {
	//CONSTRUCT MESSAGE
	message := map[string]interface{}{
		helpers.ServerActionDataMessage: data,
	}

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

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Messaging Rooms   ///////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// ServerMessage sends a server message to the specified recipients in the Room. The parameter recipients can be nil or an empty slice
// of string. In which case, the server message will be sent to all Users in the Room.
func (r *Room) ServerMessage(message interface{}, messageType int, recipients []string) error {
	if message == nil {
		return errors.New("*Room.ServerMessage() requires a message")
	}

	if serverMessageCallbackSet {
		serverMessageCallback(r, messageType, message)
	}

	return r.sendMessage(MessageTypeServer, messageType, recipients, "", message)
}

// ChatMessage sends a chat message to all Users in the Room.
func (r *Room) ChatMessage(author string, message interface{}) error {
	//REJECT INCORRECT INPUT
	if len(author) == 0 {
		return errors.New("*Room.ChatMessage() requires an author")
	} else if message == nil {
		return errors.New("*Room.ChatMessage() requires a message")
	}

	if chatMessageCallbackSet {
		chatMessageCallback(author, r, message)
	}

	return r.sendMessage(MessageTypeChat, 0, nil, author, message)
}

// DataMessage sends a data message to the specified recipients in the Room. The parameter recipients can be nil or an empty slice
// of string. In which case, the data message will be sent to all Users in the Room.
func (r *Room) DataMessage(message interface{}, recipients []string) error {
	//GET USER MAP
	userMap, err := r.GetUserMap()
	if err != nil {
		return err
	}

	//CONSTRUCT MESSAGE
	theMessage := map[string]interface{}{
		helpers.ServerActionDataMessage: message,
	}

	//SEND MESSAGE TO USERS
	if recipients == nil || len(recipients) == 0 {
		for _, u := range userMap {
			u.mux.Lock()
			for _, conn := range u.conns {
				conn.socket.WriteJSON(theMessage)
			}
			u.mux.Unlock()
		}
	} else {
		for i := 0; i < len(recipients); i++ {
			if u, ok := userMap[recipients[i]]; ok {
				u.mux.Lock()
				for _, conn := range u.conns {
					conn.socket.WriteJSON(theMessage)
				}
				u.mux.Unlock()
			}
		}
	}

	//
	return nil
}

func (r *Room) sendMessage(mt int, st int, rec []string, a string, m interface{}) error {
	//GET USER MAP
	userMap, err := r.GetUserMap()
	if err != nil {
		return err
	}

	//CONSTRUCT MESSAGE
	message := map[string]map[string]interface{}{
		helpers.ServerActionRoomMessage: make(map[string]interface{}),
	}
	// Server messages come with a sub-type
	if mt == MessageTypeServer {
		message[helpers.ServerActionRoomMessage]["s"] = st
	}
	// Non-server messages have authors
	if len(a) > 0 && mt != MessageTypeServer {
		message[helpers.ServerActionRoomMessage]["a"] = a
	}
	// The message
	message[helpers.ServerActionRoomMessage]["m"] = m

	//SEND MESSAGE TO USERS
	if rec == nil || len(rec) == 0 {
		for _, u := range userMap {
			u.mux.Lock()
			for _, conn := range u.conns {
				conn.socket.WriteJSON(message)
			}
			u.mux.Unlock()
		}
	} else {
		for i := 0; i < len(rec); i++ {
			if u, ok := userMap[rec[i]]; ok {
				u.mux.Lock()
				for _, conn := range u.conns {
					conn.socket.WriteJSON(message)
				}
				u.mux.Unlock()
			}
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   VOICE STREAMS   //////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// VoiceStream sends a voice stream from the client API to all the users in the room besides the user who is speaking.
func (r *Room) VoiceStream(userName string, userSocket *websocket.Conn, stream interface{}) {
	//GET USER MAP
	userMap, err := r.GetUserMap()
	if err != nil {
		return
	}

	//CONSTRUCT VOICE MESSAGE
	theMessage := map[string]map[string]interface{}{
		helpers.ServerActionVoiceStream: {
			"u": userName,
			"d": stream,
		},
	}

	//REMOVE SENDING USER FROM userMap
	delete(userMap, userName) // COMMENT OUT FOR ECHO TESTS

	//SEND MESSAGE TO USERS
	for _, u := range userMap {
		for _, conn := range u.conns {
			(*conn).socket.WriteJSON(theMessage)
		}
	}

	//CONSTRUCT PING MESSAGE
	pingMessage := map[string]interface{}{
		helpers.ServerActionVoicePing: nil,
	}

	//SEND PING MESSAGE TO SENDING USER
	userSocket.WriteJSON(pingMessage)

	//
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   Callback Setters   ///////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetPrivateMessageCallback sets the callback function for when a *User sends a private message to another *User.
// The function passed must have the same parameter types as the following example:
//
//    func onPrivateMessage(from *core.User, to *core.User, message interface{}) {
//	     //code...
//	 }
func SetPrivateMessageCallback(cb func(*User, *User, interface{})) {
	if !serverStarted {
		privateMessageCallback = cb
		privateMessageCallbackSet = true
	}

}

// SetChatMessageCallback sets the callback function for when a *User sends a chat message to a *Room.
// The function passed must have the same parameter types as the following example:
//
//    func onChatMessage(userName string, room *core.Room, message interface{}) {
//	     //code...
//	 }
func SetChatMessageCallback(cb func(string, *Room, interface{})) {
	if !serverStarted {
		chatMessageCallback = cb
		chatMessageCallbackSet = true
	}
}

// SetServerMessageCallback sets the callback function for when the server sends a message to a *Room.
// The function passed must have the same parameter types as the following example:
//
//    func onServerMessage(room *core.Room, messageType int, message interface{}) {
//	     //code...
//	 }
//
// The messageType value can be one of: core.ServerMessageGame, core.ServerMessageNotice,
// core.ServerMessageImportant, or a custom value you have set.
func SetServerMessageCallback(cb func(*Room, int, interface{})) {
	if !serverStarted {
		serverMessageCallback = cb
		serverMessageCallbackSet = true
	}
}
