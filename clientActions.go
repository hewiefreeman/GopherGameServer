package gopher

import (
	"github.com/hewiefreeman/GopherGameServer/users"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"github.com/hewiefreeman/GopherGameServer/actions"
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
	actionClientCustomAction = "a"
)

func clientActionHandler(action clientAction, userName *string, conn *websocket.Conn, ua *user_agent.UserAgent) (interface{}, bool, error) {
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
			return clientActionChatMessage(action.P, userName);
		case actionClientCustomAction:
			return clientCustomAction(action.P, userName, conn);
		default:
			return nil, true, errors.New("Unrecognized client action");
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   BUILT-IN CLIENT ACTIONS   ///////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionLogin(params interface{}, userName *string, conn *websocket.Conn) (interface{}, bool, error) {
	if(*userName != ""){ return nil, true, errors.New("Already logged in as '"+(*userName)+"'"); }
	//MAKE A MAP FROM PARAMS
	pMap := params.(map[string]interface{});
	//LOG IN
	user, err := users.Login(pMap["n"].(string), -1, pMap["g"].(bool), conn);
	if(err != nil){ return nil, true, err; }
	//CHANGE SOCKET'S userName
	*userName = user.Name();

	//
	return user.Name(), true, nil;
}

func clientActionLogout(userName *string) (interface{}, bool, error) {
	if(*userName == ""){ return nil, true, errors.New("Already logged out"); }
	//GET User
	user, err := users.Get(*userName);
	if(err != nil){ return nil, true, err; }
	//LOG User OUT AND RESET SOCKET'S userName
	user.LogOut();
	*userName = "";

	//
	return nil, true, nil;
}

func clientActionJoinRoom(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName == ""){ return nil, true, errors.New("Client not logged in"); }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){ return nil, true, userErr; }
	//GET ROOM NAME FROM PARAMS
	roomName := params.(string);
	//GET ROOM
	room, roomErr := rooms.Get(roomName);
	if(roomErr != nil){ return nil, true, roomErr; }
	//MAKE User JOIN THE Room
	joinErr := user.Join(room);
	if(joinErr != nil){ return nil, true, joinErr; }

	//
	return roomName, true, nil;
}

func clientActionLeaveRoom(userName *string) (interface{}, bool, error) {
	if(*userName == ""){ return nil, true, errors.New("Client not logged in"); }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		return nil, true, userErr;
	}else if(user.RoomName() == ""){
		return nil, true, errors.New("User is not in a room");
	}
	//MAKE USER LEAVE ROOM
	leaveErr := user.Leave();
	if(leaveErr != nil){ return nil, true, leaveErr; }

	//
	return nil, true, nil;
}

func clientActionCreateRoom(params interface{}, userName *string) (interface{}, bool, error) {
	if(!(*settings).UserRoomControl){
		return nil, true, errors.New("Clients do not have room control");
	}else if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		return nil, true, userErr;
	}else if(user.RoomName() != ""){
		return nil, true, errors.New("User is already in a room");
	}
	//GET PARAMS
	p := params.(map[string]interface{});
	roomName := p["n"].(string);
	roomType := p["t"].(string);
	private := p["p"].(bool);
	maxUsers := p["m"].(int);
	//MAKE THE Room
	room, roomErr := rooms.New(roomName, roomType, private, maxUsers, *userName);
	if(roomErr != nil){ return nil, true, roomErr; }
	//ADD THE User TO THE ROOM
	joinErr := user.Join(room);
	if(joinErr != nil){ return nil, true, joinErr; }

	//
	return roomName, true, nil;
}

func clientActionDeleteRoom(params interface{}, userName *string) (interface{}, bool, error) {
	if(!(*settings).UserRoomControl){
		return nil, true, errors.New("Clients do not have room control");
	}else if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}
	roomName := params.(string);
	//GET ROOM
	room, roomErr := rooms.Get(roomName);
	if(roomErr != nil){
		return nil, true, roomErr;
	}else if(room.Owner() != *userName){
		return nil, true, errors.New("User is not the owner of room '"+roomName+"'");
	}
	//DELETE ROOM
	deleteErr := room.Delete();
	if(deleteErr != nil){ return nil, true, deleteErr; }
	//
	return nil, true, nil;
}

func clientActionRoomInvite(params interface{}, userName *string) (interface{}, bool, error) {
	if(!(*settings).UserRoomControl){
		return nil, true, errors.New("Clients do not have room control");
	}else if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		return nil, true, userErr;
	}else if(user.RoomName() == ""){
		return nil, true, errors.New("User is not in a room");
	}
	//GET ROOM
	room, roomErr := rooms.Get(user.RoomName());
	if(roomErr != nil){ return nil, true, roomErr; }
	//GET PARAMS
	name := params.(string);
	//INVITE
	invErr := user.Invite(name, room);
	if(invErr != nil){ return nil, true, invErr; }
	//
	return nil, true, nil;
}

func clientActionRevokeInvite(params interface{}, userName *string) (interface{}, bool, error) {
	if(!(*settings).UserRoomControl){
		return nil, true, errors.New("Clients do not have room control");
	}else if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		return nil, true, userErr;
	}else if(user.RoomName() == ""){
		return nil, true, errors.New("User is not in a room");
	}
	//GET ROOM
	room, roomErr := rooms.Get(user.RoomName());
	if(roomErr != nil){ return nil, true, roomErr; }
	//GET PARAMS
	name := params.(string);
	//REVOKE INVITE
	revokeErr := user.RevokeInvite(name, room);
	if(revokeErr != nil){ return nil, true, revokeErr; }
	//
	return nil, true, nil;
}

func clientActionChatMessage(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName == ""){ return nil, false, errors.New("Client not logged in"); }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil || user.RoomName() == ""){ return nil, false, nil; }
	//GET ROOM
	room, roomErr := rooms.Get(user.RoomName());
	if(roomErr != nil){ return nil, false, nil; }
	//SEND CHAT MESSAGE
	room.ChatMessage(*userName, params);
	//
	return nil, false, nil;
}

func clientCustomAction(params interface{}, userName *string, conn *websocket.Conn) (interface{}, bool, error) {
	p := params.(map[string]interface{});
	action := p["a"].(string);
	data := p["d"];
	actions.HandleCustomClientAction(action, data, *userName, conn);
	return nil, false, nil;
}
