package gopher

import (
	"github.com/hewiefreeman/GopherGameServer/users"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"github.com/hewiefreeman/GopherGameServer/actions"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"github.com/hewiefreeman/GopherGameServer/database"
	"github.com/mssola/user_agent"
	"github.com/gorilla/websocket"
	"errors"
)

func clientActionHandler(action clientAction, userName *string, roomIn *rooms.Room, conn *websocket.Conn, ua *user_agent.UserAgent) (interface{}, bool, error) {
	switch _action := action.A; _action {

		case helpers.ClientActionSignup:
			return clientActionSignup(action.P);

		case helpers.ClientActionDeleteAccount:
			return clientActionDeleteAccount(action.P);

		case helpers.ClientActionChangePassword:
			return clientActionChangePassword(action.P, userName);

		case helpers.ClientActionChangeAccountInfo:
			return clientActionChangeAccountInfo(action.P, userName);

		case helpers.ClientActionLogin:
			return clientActionLogin(action.P, userName, conn);

		case helpers.ClientActionLogout:
			return clientActionLogout(userName, roomIn);

		case helpers.ClientActionJoinRoom:
			return clientActionJoinRoom(action.P, userName, roomIn);

		case helpers.ClientActionLeaveRoom:
			return clientActionLeaveRoom(userName, roomIn);

		case helpers.ClientActionCreateRoom:
			return clientActionCreateRoom(action.P, userName, roomIn);

		case helpers.ClientActionDeleteRoom:
			return clientActionDeleteRoom(action.P, userName, roomIn);

		case helpers.ClientActionRoomInvite:
			return clientActionRoomInvite(action.P, userName, roomIn);

		case helpers.ClientActionRevokeInvite:
			return clientActionRevokeInvite(action.P, userName, roomIn);

		case helpers.ClientActionChatMessage:
			return clientActionChatMessage(action.P, userName, roomIn);

		case helpers.ClientActionVoiceStream:
			return clientActionVoiceStream(action.P, userName, roomIn, conn);

		case helpers.ClientActionCustomAction:
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

func clientActionSignup(params interface{}) (interface{}, bool, error) {
	//GET ITEMS FROM PARAMS
	pMap := params.(map[string]interface{});
	customCols := pMap["c"].(map[string]interface{});
	userName := pMap["n"].(string);
	pass := pMap["p"].(string);
	//SIGN CLIENT UP
	signupErr := database.SignUpClient(userName, pass, customCols);
	if(signupErr != nil){ return nil, true, signupErr; }

	//
	return nil, true, nil;
}

func clientActionDeleteAccount(params interface{}) (interface{}, bool, error) {
	//GET ITEMS FROM PARAMS
	pMap := params.(map[string]interface{});
	customCols := pMap["c"].(map[string]interface{});
	userName := pMap["n"].(string);
	pass := pMap["p"].(string);
	//DELETE ACCOUNT
	deleteErr := database.DeleteAccount(userName, pass, customCols);
	if(deleteErr != nil){ return nil, true, deleteErr; }
	//LOG USER OUT IF ONLINE
	users.DropUser(userName);

	//
	return nil, true, nil;
}

func clientActionChangePassword(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName != ""){
		return nil, true, errors.New("You must be logged in to change your password");
	}
	//GET ITEMS FROM PARAMS
	pMap := params.(map[string]interface{});
	customCols := pMap["c"].(map[string]interface{});
	pass := pMap["p"].(string);
	//CHANGE PASSWORD
	changeErr := database.ChangePassword(*userName, pass, customCols);
	if(changeErr != nil){ return nil, true, changeErr; }

	//
	return nil, true, nil;
}

func clientActionChangeAccountInfo(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName != ""){
		return nil, true, errors.New("You must be logged in to change your account info");
	}
	//GET ITEMS FROM PARAMS
	pMap := params.(map[string]interface{});
	customCols := pMap["c"].(map[string]interface{});
	pass := pMap["p"].(string);
	//CHANGE ACCOUNT INFO
	changeErr := database.ChangeAccountInfo(*userName, pass, customCols);
	if(changeErr != nil){ return nil, true, changeErr; }

	//
	return nil, true, nil;
}

func clientActionLogin(params interface{}, userName *string, conn *websocket.Conn) (interface{}, bool, error) {
	if(*userName != ""){ return nil, true, errors.New("Already logged in as '"+(*userName)+"'"); }
	//MAKE A MAP FROM PARAMS
	pMap := params.(map[string]interface{});
	name := pMap["n"].(string);
	pass := pMap["p"].(string);
	guest := pMap["g"].(bool);
	customCols := pMap["c"].(map[string]interface{});
	//LOG IN
	var dbIndex int;
	var user users.User;
	var err error;
	if((*settings).EnableSqlFeatures){
		dbIndex, err = database.LoginClient(name, pass, customCols);
		if(err != nil){ return nil, true, err; }
		user, err = users.Login(name, dbIndex, guest, conn);
	}else{
		user, err = users.Login(name, -1, guest, conn);
	}
	if(err != nil){ return nil, true, err; }
	//CHANGE SOCKET'S userName
	*userName = user.Name();

	//
	return nil, false, nil;
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
	return nil, false, nil;
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
	return nil, false, nil;
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
	return nil, false, nil;
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
	if(rType.ServerOnly()){ return nil, true, errors.New("Only the server can manipulate that type of room"); }
	//DELETE ROOM
	deleteErr := room.Delete();
	if(deleteErr != nil){ return nil, true, deleteErr; }
	//
	if(roomIn.Name() == roomName){ *roomIn = room; }

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
	if(rType.ServerOnly()){ return nil, true, errors.New("Only the server can manipulate that type of room"); }
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
	if(rType.ServerOnly()){ return nil, true, errors.New("Only the server can manipulate that type of room"); }
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
	//CHECK IF VOICE CHAT ROOM
	rType := rooms.GetRoomTypes()[(*roomIn).Type()];
	if(!rType.VoiceChatEnabled()){ return nil, false, nil; }
	//SEND VOICE STREAM
	(*roomIn).VoiceStream(*userName, conn, params);
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
	(*roomIn).ChatMessage(*userName, params);
	//
	return nil, false, nil;
}
