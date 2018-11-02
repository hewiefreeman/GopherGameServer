package rooms

import (
	"github.com/gorilla/websocket"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   VOICE MESSAGES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Room) VoiceStream(userName string, userSocket *websocket.Conn, stream interface{}){
	//GET USER MAP
	userMap, err := r.GetUserMap();
	if(err != nil){ return; }

	//CONSTRUCT PING MESSAGE
	pingMessage := make(map[string]interface{});
	pingMessage["vp"] = nil;

	userSocket.WriteJSON(pingMessage);

	//CONSTRUCT MESSAGE
	theMessage := make(map[string]interface{});
	theMessage["v"] = make(map[string]interface{}); // Voice streams are labeled "v"
	theMessage["v"].(map[string]interface{})["u"] = userName;
	theMessage["v"].(map[string]interface{})["d"] = stream;

	//SEND MESSAGE TO USERS
	for _, u := range userMap {
		/*if(k != userName){*/u.socket.WriteJSON(theMessage);//}
	}

	//
	return;
}
