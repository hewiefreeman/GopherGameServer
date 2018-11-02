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
	actionClientVoiceStream = "v"
	actionClientCustomAction = "a"
)

func clientActionHandler(action clientAction, userName *string, roomIn *rooms.Room, conn *websocket.Conn, ua *user_agent.UserAgent) (interface{}, bool, error) {
	switch _action := action.A; _action {
		case actionClientLogin:
			return clientActionLogin(action.P, userName, conn);
		case actionClientLogout:
			return clientActionLogout(userName, roomIn);
		case actionClientJoinRoom:
			return clientActionJoinRoom(action.P, userName, roomIn);
		case actionClientLeaveRoom:
			return clientActionLeaveRoom(userName, roomIn);
		case actionClientCreateRoom:
			return clientActionCreateRoom(action.P, userName, roomIn);
		case actionClientDeleteRoom:
			return clientActionDeleteRoom(action.P, userName, roomIn);
		case actionClientRoomInvite:
			return clientActionRoomInvite(action.P, userName, roomIn);
		case actionClientRevokeInvite:
			return clientActionRevokeInvite(action.P, userName, roomIn);
		case actionClientChatMessage:
			return clientActionChatMessage(action.P, userName, roomIn);
		case actionClientVoiceStream:
			return clientActionVoiceStream(action.P, userName, roomIn, conn);
		case actionClientCustomAction:
			return clientCustomAction(action.P, userName, conn);
		default:
			return nil, true, errors.New("Unrecognized client action");
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CUSTOM CLIENT ACTIONS   /////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientCustomAction(params interface{}, userName *string, conn *websocket.Conn) (interface{}, bool, error) {
	p := params.(map[string]interface{});
	action := p["a"].(string);
	data := p["d"];
	actions.HandleCustomClientAction(action, data, *userName, conn);
	return nil, false, nil;
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

func clientActionLogout(userName *string, roomIn *rooms.Room) (interface{}, bool, error) {
	if(*userName == ""){ return nil, true, errors.New("Already logged out"); }
	//GET User
	user, err := users.Get(*userName);
	if(err != nil){
		*userName = "";
		*roomIn = rooms.Room{};
		return nil, true, err;
	}
	//LOG User OUT AND RESET SOCKET'S userName
	user.LogOut();
	*userName = "";
	*roomIn = rooms.Room{};

	//
	return nil, true, nil;
}

func clientActionJoinRoom(params interface{}, userName *string, roomIn *rooms.Room) (interface{}, bool, error) {
	if(*userName == ""){ return nil, true, errors.New("Client not logged in"); }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		*userName = "";
		*roomIn = rooms.Room{};
		return nil, true, userErr;
	}
	//GET ROOM NAME FROM PARAMS
	roomName := params.(string);
	//GET ROOM
	room, roomErr := rooms.Get(roomName);
	if(roomErr != nil){ return nil, true, roomErr; }
	//MAKE User JOIN THE Room
	joinErr := user.Join(room);
	if(joinErr != nil){ return nil, true, joinErr; }
	//
	*roomIn = room;

	//
	return roomName, true, nil;
}

func clientActionLeaveRoom(userName *string, roomIn *rooms.Room) (interface{}, bool, error) {
	if(*userName == ""){ return nil, true, errors.New("Client not logged in"); }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		*userName = "";
		*roomIn = rooms.Room{};
		return nil, true, userErr;
	}else if(user.RoomName() == ""){
		*roomIn = rooms.Room{};
		return nil, true, errors.New("User is not in a room");
	}
	//MAKE USER LEAVE ROOM
	leaveErr := user.Leave();
	if(leaveErr != nil){
		*roomIn = rooms.Room{};
		return nil, true, leaveErr;
	}
	//
	*roomIn = rooms.Room{};

	//
	return nil, true, nil;
}

func clientActionCreateRoom(params interface{}, userName *string, roomIn *rooms.Room) (interface{}, bool, error) {
	if(!(*settings).UserRoomControl){
		return nil, true, errors.New("Clients do not have room control");
	}else if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}
	//GET PARAMS
	p := params.(map[string]interface{});
	roomName := p["n"].(string);
	roomType := p["t"].(string);
	private := p["p"].(bool);
	maxUsers := p["m"].(int);
	//
	if rType, ok := rooms.GetRoomTypes()[roomType]; !ok {
		return nil, true, errors.New("Invalid room type");
	}else if(rType.ServerOnly()){
		return nil, true, errors.New("Only the server can manipulate that type of room");
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		*userName = "";
		*roomIn = rooms.Room{};
		return nil, true, userErr;
	}else if(user.RoomName() != ""){
		return nil, true, errors.New("User is already in a room");
	}
	//MAKE THE Room
	room, roomErr := rooms.New(roomName, roomType, private, maxUsers, *userName);
	if(roomErr != nil){ return nil, true, roomErr; }
	//ADD THE User TO THE ROOM
	joinErr := user.Join(room);
	if(joinErr != nil){ return nil, true, joinErr; }
	//
	*roomIn = room;

	//
	return roomName, true, nil;
}

func clientActionDeleteRoom(params interface{}, userName *string, roomIn *rooms.Room) (interface{}, bool, error) {
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
	//
	rType := rooms.GetRoomTypes()[room.Type()];
	if(rType.ServerOnly()){
		return nil, true, errors.New("Only the server can manipulate that type of room");
	}
	//DELETE ROOM
	deleteErr := room.Delete();
	if(deleteErr != nil){ return nil, true, deleteErr; }
	//
	if(roomIn.Name() == roomName){
		*roomIn = room;
	}

	return nil, true, nil;
}

func clientActionRoomInvite(params interface{}, userName *string, roomIn *rooms.Room) (interface{}, bool, error) {
	if(!(*settings).UserRoomControl){
		return nil, true, errors.New("Clients do not have room control");
	}else if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		*userName = "";
		*roomIn = rooms.Room{};
		return nil, true, userErr;
	}else if(user.RoomName() == ""){
		*roomIn = rooms.Room{};
		return nil, true, errors.New("User is not in a room");
	}
	//
	rType := rooms.GetRoomTypes()[(*roomIn).Type()];
	if(rType.ServerOnly()){
		return nil, true, errors.New("Only the server can manipulate that type of room");
	}
	//GET PARAMS
	name := params.(string);
	//INVITE
	invErr := user.Invite(name, *roomIn);
	if(invErr != nil){ return nil, true, invErr; }
	//
	return nil, true, nil;
}

func clientActionRevokeInvite(params interface{}, userName *string, roomIn *rooms.Room) (interface{}, bool, error) {
	if(!(*settings).UserRoomControl){
		return nil, true, errors.New("Clients do not have room control");
	}else if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){
		*userName = "";
		*roomIn = rooms.Room{};
		return nil, true, userErr;
	}else if(user.RoomName() == ""){
		*roomIn = rooms.Room{};
		return nil, true, errors.New("User is not in a room");
	}
	//
	rType := rooms.GetRoomTypes()[(*roomIn).Type()];
	if(rType.ServerOnly()){
		return nil, true, errors.New("Only the server can manipulate that type of room");
	}
	//GET PARAMS
	name := params.(string);
	//REVOKE INVITE
	revokeErr := user.RevokeInvite(name, *roomIn);
	if(revokeErr != nil){ return nil, true, revokeErr; }
	//
	return nil, true, nil;
}

func clientActionVoiceStream(params interface{}, userName *string, roomIn *rooms.Room, conn *websocket.Conn) (interface{}, bool, error) {
	if(*userName == ""){ return nil, false, errors.New("Client not logged in"); }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil || user.RoomName() == ""){
		*userName = "";
		*roomIn = rooms.Room{};
		return nil, false, nil;
	}
	//SEND VOICE STREAM
	go (*roomIn).VoiceStream(*userName, conn, params);
	//
	return nil, false, nil;
}

func clientActionChatMessage(params interface{}, userName *string, roomIn *rooms.Room) (interface{}, bool, error) {
	if(*userName == ""){ return nil, false, errors.New("Client not logged in"); }
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil || user.RoomName() == ""){
		*userName = "";
		*roomIn = rooms.Room{};
		return nil, false, nil;
	}
	//SEND CHAT MESSAGE
	go (*roomIn).ChatMessage(*userName, params);
	//
	return nil, false, nil;
}
