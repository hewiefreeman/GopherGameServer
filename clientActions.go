package gopher

import (
	"github.com/hewiefreeman/GopherGameServer/users"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"github.com/mssola/user_agent"
	"github.com/gorilla/websocket"
	"errors"
)

const (
	actionClientLogin = "li"
	actionClientLogout = "lo"
	actionClientJoinRoom = "j"
	actionClientLeaveRoom = "lr"
	actionClientChatMessage = "c"
)

func clientActionHandler(action clientAction, userName *string, roomIn *rooms.Room, conn *websocket.Conn, ua *user_agent.UserAgent) (interface{}, error, bool) {
	switch _action := action.A; _action {
		case actionClientLogin:
			return clientActionLogin(action.P, userName, conn);
		case actionClientLogout:
			return clientActionLogout(userName, roomIn);
		case actionClientJoinRoom:
			return clientActionJoinRoom(action.P, userName, roomIn, conn);
		case actionClientLeaveRoom:
			return clientActionLeaveRoom(userName, roomIn);
		default:
			return nil, errors.New("Unrecognized client action"), true;
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   BUILT-IN CLIENT ACTIONS   ///////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionLogin(params interface{}, userName *string, conn *websocket.Conn) (interface{}, error, bool) {
	if(*userName != ""){ return nil, errors.New("Already logged in as '"+(*userName)+"'"), true; }
	//MAKE A MAP FROM PARAMS
	pMap := params.(map[string]interface{});
	//LOG IN
	user, err := users.Login(pMap["n"].(string), -1, pMap["g"].(bool), conn);
	if(err != nil){ return nil, err, true }
	//CHANGE SOCKET'S userName
	*userName = user.Name();

	//
	return user.Name(), nil, true;
}

func clientActionLogout(userName *string, roomIn *rooms.Room) (interface{}, error, bool) {
	if(*userName == ""){ return nil, errors.New("Already logged out"), true; }
	//GET User
	user, err := users.Get(*userName);
	if(err != nil){ return nil, err, true }
	//LOG User OUT AND RESET SOCKET'S userName & roomIn
	user.Logout();
	*userName = "";
	*roomIn = rooms.Room{};

	//
	return nil, nil, true;
}

func clientActionJoinRoom(params interface{}, userName *string, roomIn *rooms.Room, conn *websocket.Conn) (interface{}, error, bool) {
	if(*userName == ""){ return nil, errors.New("Client not logged in"), true; }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){ return nil, userErr, true; }
	//GET ROOM NAME FROM PARAMS
	roomName := params.(string);
	//GET ROOM
	room, roomErr := rooms.Get(roomName);
	if(roomErr != nil){ return nil, roomErr, true; }
	//MAKE User JOIN THE Room
	joinErr := user.Join(room);
	if(joinErr != nil){ return nil, joinErr, true; }
	//CHANGE SOCKET'S roomIn
	*roomIn = room;

	//
	return roomName, nil, true;
}

func clientActionLeaveRoom(userName *string, roomIn *rooms.Room) (interface{}, error, bool) {
	if(*userName == ""){
		return nil, errors.New("Client not logged in"), true;
	}else if(roomIn.Name() == ""){
		return nil, errors.New("User is not in a room"), true;
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){ return nil, userErr, true; }
	//MAKE USER LEAVE ROOM
	leaveErr := user.Leave();
	if(leaveErr != nil){ return nil, leaveErr, true; }
	//RESET SOCKET'S roomIn
	*roomIn = rooms.Room{};

	//
	return roomName, nil, true;
}

func actionClientChatMessage(params interface{}, userName *string, roomIn *rooms.Room) (interface{}, error, bool) {
	roomIn.ChatMessage(*userName, params);
	return nil, nil, false;
}
