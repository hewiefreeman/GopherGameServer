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
	actionClientCreateRoom = "r"
	actionClientDeleteRoom = "rd"
	actionClientRoomInvite = "i"
	actionClientRevokeInvite = "ri"
	actionClientChatMessage = "c"
)

func clientActionHandler(action clientAction, userName *string, conn *websocket.Conn, ua *user_agent.UserAgent) (interface{}, error, bool) {
	switch _action := action.A; _action {
		case actionClientLogin:
			return clientActionLogin(action.P, userName, conn);
		case actionClientLogout:
			return clientActionLogout(userName);
		case actionClientJoinRoom:
			return clientActionJoinRoom(action.P, userName);
		case actionClientLeaveRoom:
			return clientActionLeaveRoom(userName);
		case actionClientCreateRoom:
			return clientActionCreateRoom(action.P, userName);
		case actionClientDeleteRoom:
			return clientActionDeleteRoom(action.P, userName);
		case actionClientRoomInvite:
			return clientActionRoomInvite(action.P, userName);
		case actionClientRevokeInvite:
			return clientActionRevokeInvite(action.P, userName);
		case actionClientChatMessage:
			return clientActionChatMessage(action.P);
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

func clientActionLogout(userName *string) (interface{}, error, bool) {
	if(*userName == ""){ return nil, errors.New("Already logged out"), true; }
	//GET User
	user, err := users.Get(*userName);
	if(err != nil){ return nil, err, true }
	//LOG User OUT AND RESET SOCKET'S userName
	user.Logout();
	*userName = "";

	//
	return nil, nil, true;
}

func clientActionJoinRoom(params interface{}, userName *string) (interface{}, error, bool) {
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

	//
	return roomName, nil, true;
}

func clientActionLeaveRoom(userName *string) (interface{}, error, bool) {
	if(*userName == ""){ return nil, errors.New("Client not logged in"), true; }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		return nil, userErr, true;
	}else if(user.RoomName() == ""){
		return nil, errors.New("User is not in a room"), true;
	}
	//MAKE USER LEAVE ROOM
	leaveErr := user.Leave();
	if(leaveErr != nil){ return nil, leaveErr, true; }

	//
	return roomName, nil, true;
}

func clientActionCreateRoom(params interface{}, userName *string) (interface{}, error, bool) {
	if(*userName == ""){ return nil, errors.New("Client not logged in"), true; }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		return nil, userErr, true;
	}else if(user.RoomName() != ""){
		return nil, errors.New("User is already in a room"), true;
	}
	//GET PARAMS
	p := params.(map[string]interface{});
	roomName := p["n"].(string);
	roomType := p["t"].(string);
	private := p["p"].(bool);
	maxUsers := p["m"].(int);
	//MAKE THE Room
	room, roomErr := rooms.New(roomName, roomType, private, maxUsers, *userName);
	if(roomErr != nil){ return nil, roomErr, true; }
	//ADD THE User TO THE ROOM
	joinErr := user.Join(room);
	if(joinErr != nil){ return nil, joinErr, true; }

	//
	return roomName, nil, true;
}

func clientActionDeleteRoom(params interface{}, userName *string) (interface{}, error, bool) {
	if(*userName == ""){ return nil, errors.New("Client not logged in"), true; }
	roomName := params.(string);
	//GET ROOM
	room, roomErr := rooms.Get(roomName);
	if(roomErr != nil){
		return nil, roomErr, true;
	}else if(room.Owner() != *userName){
		return nil, errors.New("User is not the owner of room '"+roomName+"'"), true;
	}
	//DELETE ROOM
	deleteErr := room.Delete();
	if(deleteErr != nil){ return nil, deleteErr, true; }
	//
	return nil, nil, true;
}

func clientActionRoomInvite(params interface{}, userName *string) (interface{}, error, bool) {
	if(*userName == ""){ return nil, errors.New("Client not logged in"), true; }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		return nil, userErr, true;
	}else if(user.RoomName() == ""){
		return nil, errors.New("User is not in a room"), true;
	}
	//GET ROOM
	room, roomErr := rooms.Get(user.RoomName());
	if(roomErr != nil){ return nil, roomErr, true; }
	//GET PARAMS
	userName := params.(string);
	//INVITE
	invErr := user.Invite(userName, room);
	if(invErr != nil){ return nil, invErr, true; }
	//
	return nil, nil, true;
}

func clientActionRevokeInvite(params interface{}, userName *string) (interface{}, error, bool) {
	if(*userName == ""){ return nil, errors.New("Client not logged in"), true; }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		return nil, userErr, true;
	}else if(user.RoomName() == ""){
		return nil, errors.New("User is not in a room"), true;
	}
	//GET ROOM
	room, roomErr := rooms.Get(user.RoomName());
	if(roomErr != nil){ return nil, roomErr, true; }
	//GET PARAMS
	userName := params.(string);
	//REVOKE INVITE
	revokeErr := user.RevokeInvite(userName, room);
	if(revokeErr != nil){ return nil, revokeErr, true; }
	//
	return nil, nil, true;
}

func clientActionChatMessage(params interface{}, userName *string) (interface{}, error, bool) {
	if(*userName == ""){ return nil, errors.New("Client not logged in"), true; }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil || user.RoomName() == ""){ return nil, nil, false; }
	//GET ROOM
	room, roomErr := rooms.Get(user.RoomName());
	if(roomErr != nil){ return nil, nil, false; }
	//SEND CHAT MESSAGE
	room.ChatMessage(*userName, params);
	//
	return nil, nil, false;
}
