package rooms

import (
	"github.com/gorilla/websocket"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   VOICE STREAMS   //////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Room) VoiceStream(userName string, userSocket *websocket.Conn, stream interface{}){
	//GET USER MAP
	userMap, err := r.GetUserMap();
	if(err != nil){ return; }

	//CONSTRUCT VOICE MESSAGE
	theMessage := make(map[string]interface{});
	theMessage["v"] = make(map[string]interface{}); // Voice streams are labeled "v"
	theMessage["v"].(map[string]interface{})["u"] = userName;
	theMessage["v"].(map[string]interface{})["d"] = stream;

	//REMOVE SENDING USER FROM userMap
	//delete(userMap, userName); // COMMENTED OUT FOR ECHO TESTS

	//SEND MESSAGE TO USERS
	for _, u := range userMap { u.socket.WriteJSON(theMessage); }

	//CONSTRUCT PING MESSAGE
	pingMessage := make(map[string]interface{});
	pingMessage["vp"] = nil;

	//SEND PING MESSAGE TO SENDING USER
	userSocket.WriteJSON(pingMessage);

	//
	return;
}
