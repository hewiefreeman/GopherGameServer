package rooms

import (
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/helpers"
)

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
	theMessage := make(map[string]interface{})
	theMessage[helpers.ServerActionVoiceStream] = make(map[string]interface{})
	theMessage[helpers.ServerActionVoiceStream].(map[string]interface{})["u"] = userName
	theMessage[helpers.ServerActionVoiceStream].(map[string]interface{})["d"] = stream

	//REMOVE SENDING USER FROM userMap
	delete(userMap, userName); // COMMENT OUT FOR ECHO TESTS

	//SEND MESSAGE TO USERS
	for _, u := range userMap {
		u.socket.WriteJSON(theMessage)
	}

	//CONSTRUCT PING MESSAGE
	pingMessage := make(map[string]interface{})
	pingMessage[helpers.ServerActionVoicePing] = nil

	//SEND PING MESSAGE TO SENDING USER
	userSocket.WriteJSON(pingMessage)

	//
	return
}
