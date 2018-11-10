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
	"fmt"
)

const (
	errorIncorrectFormat = "Incorrect data format"
	errorIncorrectFormatName = "Incorrect data format for user name"
	errorIncorrectFormatPass = "Incorrect data format for password"
	errorIncorrectFormatNewPass = "Incorrect data format for new password"
	errorIncorrectFormatAction = "Incorrect data format for action"
	errorIncorrectFormatCols = "Incorrect data format for custom columns"
	errorIncorrectFormatRemember = "Incorrect data format for remember me"
	errorIncorrectFormatGuest = "Incorrect data format for guest"
	errorIncorrectFormatRoomName = "Incorrect data format for room name"
	errorIncorrectFormatRoomType = "Incorrect data format for room type"
	errorIncorrectFormatPrivateRoom = "Incorrect data format for private room"
	errorIncorrectFormatMaxRoomUsers = "Incorrect data format for max room users"
)

func clientActionHandler(action clientAction, userName *string, roomIn *rooms.Room, conn *websocket.Conn, ua *user_agent.UserAgent,
					deviceTag *string, devicePass *string, deviceUserID *int) (interface{}, bool, error) {
	switch _action := action.A; _action {

		// DATABASE

		case helpers.ClientActionSignup:
			return clientActionSignup(action.P);
		case helpers.ClientActionDeleteAccount:
			return clientActionDeleteAccount(action.P);
		case helpers.ClientActionChangePassword:
			return clientActionChangePassword(action.P, userName);
		case helpers.ClientActionChangeAccountInfo:
			return clientActionChangeAccountInfo(action.P, userName);

		// LOGIN/LOGOUT

		case helpers.ClientActionLogin:
			return clientActionLogin(action.P, userName, deviceTag, devicePass, deviceUserID, conn);
		case helpers.ClientActionLogout:
			return clientActionLogout(userName, roomIn, deviceTag, devicePass, deviceUserID);

		// ROOM ACTIONS

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

		// CHAT+VOICE

		case helpers.ClientActionChatMessage:
			return clientActionChatMessage(action.P, userName, roomIn);
		case helpers.ClientActionVoiceStream:
			return clientActionVoiceStream(action.P, userName, roomIn, conn);

		// CHANGE STATUS

		case helpers.ClientActionChangeStatus:
			return clientActionChangeStatus(action.P, userName);

		// CUSTOM ACTIONS

		case helpers.ClientActionCustomAction:
			return clientCustomAction(action.P, userName, conn);

		// FRIENDING

		case helpers.ClientActionFriendRequest:
			return clientActionFriendRequest(action.P, userName);
		case helpers.ClientActionAcceptFriend:
			return clientActionAcceptFriend(action.P, userName);
		case helpers.ClientActionDeclineFriend:
			return clientActionDeclineFriend(action.P, userName);
		case helpers.ClientActionRemoveFriend:
			return clientActionRemoveFriend(action.P, userName);

		// INVALID CLIENT ACTION

		default:
			return nil, true, errors.New("Unrecognized client action");
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CUSTOM CLIENT ACTIONS   /////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientCustomAction(params interface{}, userName *string, conn *websocket.Conn) (interface{}, bool, error) {
	var ok bool;
	var pMap map[string]interface{};
	var action string;
	if pMap, ok = params.(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormat); }
	if action, ok = pMap["a"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatAction); }
	actions.HandleCustomClientAction(action, pMap["d"], *userName, conn);
	return nil, false, nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHANGE USER STATUS   ////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionChangeStatus(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName != ""){ return nil, true, errors.New("You must be logged in to change your status"); }
	user, userErr := users.Get(*userName);
	if(userErr != nil){ return nil, true, userErr; }
	//GET PARAMS
	var ok bool;
	var status int;

	if status, ok = params.(int); !ok { return nil, true, errors.New(errorIncorrectFormat); }
	//
	statusErr := user.SetStatus(status);
	if(statusErr != nil){ return nil, true, statusErr; }
	//
	return status, true, nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ACCOUNT/DATABASE ACTIONS   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionSignup(params interface{}) (interface{}, bool, error) {
	if(!(*settings).EnableSqlFeatures){
		return nil, true, errors.New("SQL Features are not enabled");
	}
	//GET ITEMS FROM PARAMS
	var ok bool;
	var pMap map[string]interface{};
	var customCols map[string]interface{};
	var userName string;
	var pass string;
	if pMap, ok = params.(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormat); }
	if(pMap["c"] != nil){
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormatCols); }
	}
	if userName, ok = pMap["n"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatName); }
	if pass, ok = pMap["p"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatPass); }
	//SIGN CLIENT UP
	signupErr := database.SignUpClient(userName, pass, customCols);
	if(signupErr != nil){ return nil, true, signupErr; }

	//
	return nil, true, nil;
}

func clientActionDeleteAccount(params interface{}) (interface{}, bool, error) {
	if(!(*settings).EnableSqlFeatures){
		return nil, true, errors.New("SQL Features are not enabled");
	}
	//GET ITEMS FROM PARAMS
	var ok bool;
	var pMap map[string]interface{};
	var customCols map[string]interface{};
	var userName string;
	var pass string;
	if pMap, ok = params.(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormat); }
	if(pMap["c"] != nil){
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormatCols); }
	}
	if userName, ok = pMap["n"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatName); }
	if pass, ok = pMap["p"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatPass); }

	//CHECK IF USER IS ONLINE
	_, err := users.Get(userName);
	if(err == nil){ return nil, true, errors.New("The User must be logged off to delete their account."); }

	//DELETE ACCOUNT
	deleteErr := database.DeleteAccount(userName, pass, customCols);
	if(deleteErr != nil){ return nil, true, deleteErr; }
	fmt.Println("got through delete...?");
	//
	return nil, true, nil;
}

func clientActionChangePassword(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName == ""){
		return nil, true, errors.New("You must be logged in to change your password");
	}else if(!(*settings).EnableSqlFeatures){
		return nil, true, errors.New("SQL Features are not enabled");
	}
	//GET ITEMS FROM PARAMS
	var ok bool;
	var pMap map[string]interface{};
	var customCols map[string]interface{};
	var pass string;
	var newPass string;
	if pMap, ok = params.(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormat); }
	if(pMap["c"] != nil){
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormatCols); }
	}
	if pass, ok = pMap["p"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatPass); }
	if newPass, ok = pMap["n"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatNewPass); }
	//CHANGE PASSWORD
	changeErr := database.ChangePassword(*userName, pass, newPass, customCols);
	if(changeErr != nil){ return nil, true, changeErr; }

	//
	return nil, true, nil;
}

func clientActionChangeAccountInfo(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName == ""){
		return nil, true, errors.New("You must be logged in to change your account info");
	}else if(!(*settings).EnableSqlFeatures){
		return nil, true, errors.New("SQL Features are not enabled");
	}
	//GET ITEMS FROM PARAMS
	var ok bool;
	var pMap map[string]interface{};
	var customCols map[string]interface{};
	var pass string;
	if pMap, ok = params.(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormat); }
	if(pMap["c"] != nil){
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormatCols); }
	}
	if pass, ok = pMap["p"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatPass); }
	//CHANGE ACCOUNT INFO
	changeErr := database.ChangeAccountInfo(*userName, pass, customCols);
	if(changeErr != nil){ return nil, true, changeErr; }

	//
	return nil, true, nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOGIN+LOGOUT ACTIONS   //////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionLogin(params interface{}, userName *string, deviceTag *string, devicePass *string, deviceUserID *int, conn *websocket.Conn) (interface{}, bool, error) {
	if(*userName != ""){ return nil, true, errors.New("Already logged in as '"+(*userName)+"'"); }
	//MAKE A MAP FROM PARAMS
	var ok bool;
	var pMap map[string]interface{};
	var name string;
	var pass string;
	var remMe bool = false;
	var guest bool = false;
	var customCols map[string]interface{};
	if pMap, ok = params.(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormat); }
	if name, ok = pMap["n"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatName); }
	if((*settings).EnableSqlFeatures){
		if pass, ok = pMap["p"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatPass); }
		if((*settings).RememberMe){
			if remMe, ok = pMap["r"].(bool); !ok { return nil, true, errors.New(errorIncorrectFormatRemember); }
		}
	}
	if(pMap["g"] != nil){
		if guest, ok = pMap["g"].(bool); !ok { return nil, true, errors.New(errorIncorrectFormatGuest); }
	}
	if(pMap["c"] != nil){
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormatCols); }
	}
	//LOG IN
	var dbIndex int;
	var user users.User;
	var uName string;
	var dPass string;
	var err error;
	if((*settings).EnableSqlFeatures){
		uName, dbIndex, dPass, err = database.LoginClient(name, pass, *deviceTag, remMe, customCols);
		if(err != nil){ return nil, true, err; }
		user, err = users.Login(uName, dbIndex, dPass, guest, remMe, conn);
	}else{
		user, err = users.Login(name, -1, "", guest, false, conn);
	}
	if(err != nil){ return nil, true, err; }
	//CHANGE SOCKET'S userName
	*userName = user.Name();
	*devicePass = dPass;
	*deviceUserID = dbIndex;

	//
	return nil, false, nil;
}

func clientActionLogout(userName *string, roomIn *rooms.Room, deviceTag *string, devicePass *string, deviceUserID *int) (interface{}, bool, error) {
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
	//REMOVE AUTO-LOG IF ANY
	if((*settings).EnableSqlFeatures && (*settings).RememberMe){
		database.RemoveAutoLog(*deviceUserID, *deviceTag);
	}
	//
	*userName = "";
	*roomIn = rooms.Room{};
	*devicePass = "";
	*deviceUserID = 0;

	//
	return nil, false, nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ROOM ACTIONS   //////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

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
	var ok bool;
	var roomName string;
	if roomName, ok = params.(string); !ok { return nil, true, errors.New(errorIncorrectFormat); }
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
	var ok bool;
	var pMap map[string]interface{};
	var roomName string;
	var roomType string;
	var private bool;
	var maxUsers int;
	if pMap, ok = params.(map[string]interface{}); !ok { return nil, true, errors.New(errorIncorrectFormat); }
	if roomName, ok = pMap["n"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatRoomName); }
	if roomType, ok = pMap["t"].(string); !ok { return nil, true, errors.New(errorIncorrectFormatRoomType); }
	if private, ok = pMap["t"].(bool); !ok { return nil, true, errors.New(errorIncorrectFormatPrivateRoom); }
	if maxUsers, ok = pMap["m"].(int); !ok { return nil, true, errors.New(errorIncorrectFormatMaxRoomUsers); }
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
	//GET PARAMS
	var ok bool;
	var roomName string;
	if roomName, ok = params.(string); !ok { return nil, true, errors.New(errorIncorrectFormat); }
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
	var ok bool;
	var name string;
	if name, ok = params.(string); !ok { return nil, true, errors.New(errorIncorrectFormat); }
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
	var ok bool;
	var name string;
	if name, ok = params.(string); !ok { return nil, true, errors.New(errorIncorrectFormat); }
	//REVOKE INVITE
	revokeErr := user.RevokeInvite(name, *roomIn);
	if(revokeErr != nil){ return nil, true, revokeErr; }
	//
	return nil, true, nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHAT+VOICE ACTIONS   ////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionVoiceStream(params interface{}, userName *string, roomIn *rooms.Room, conn *websocket.Conn) (interface{}, bool, error) {
	if(*userName == ""){ return nil, false, nil; }
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
	if(*userName == ""){ return nil, false, nil; }
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

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   FRIENDING ACTIONS   /////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionFriendRequest(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}else if(!(*settings).EnableSqlFeatures){
		return nil, true, errors.New("SQL Features are not enabled");
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){ return nil, true, errors.New("Client not logged in"); }
	//GET PARAMS AS A MAP
	var ok bool;
	var friendName string;

	if friendName, ok = params.(string); !ok { return nil, true, errors.New(errorIncorrectFormat); }

	requestErr := user.FriendRequest(friendName);
	if(userErr != nil){ return nil, true, requestErr; }

	//
	return friendName, true, nil;
}

func clientActionAcceptFriend(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}else if(!(*settings).EnableSqlFeatures){
		return nil, true, errors.New("SQL Features are not enabled");
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){ return nil, true, errors.New("Client not logged in"); }
	//GET PARAMS AS A MAP
	var ok bool;
	var friendName string;

	if friendName, ok = params.(string); !ok { return nil, true, errors.New(errorIncorrectFormat); }

	acceptErr := user.AcceptFriendRequest(friendName);
	if(acceptErr != nil){ return nil, true, acceptErr; }

	//GET FRIEND'S status
	var status int;
	friend, friendErr := users.Get(friendName);
	if(friendErr != nil){
		status = users.StatusOffline;
	}else{
		status = friend.Status();
	}

	//MAKE RESPONSE
	responseMap := make(map[string]interface{});
	responseMap["n"] = friendName;
	responseMap["s"] = status;

	//
	return responseMap, true, nil;
}

func clientActionDeclineFriend(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}else if(!(*settings).EnableSqlFeatures){
		return nil, true, errors.New("SQL Features are not enabled");
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){ return nil, true, errors.New("Client not logged in"); }
	//GET PARAMS AS A MAP
	var ok bool;
	var friendName string;

	if friendName, ok = params.(string); !ok { return nil, true, errors.New(errorIncorrectFormat); }

	declineErr := user.DeclineFriendRequest(friendName);
	if(declineErr != nil){ return nil, true, declineErr; }

	//
	return friendName, true, nil;
}

func clientActionRemoveFriend(params interface{}, userName *string) (interface{}, bool, error) {
	if(*userName == ""){
		return nil, true, errors.New("Client not logged in");
	}else if(!(*settings).EnableSqlFeatures){
		return nil, true, errors.New("SQL Features are not enabled");
	}
	//GET User
	user, userErr := users.Get(*userName);
	if(userErr != nil){ return nil, true, errors.New("Client not logged in"); }
	//GET PARAMS AS A MAP
	var ok bool;
	var friendName string;

	if friendName, ok = params.(string); !ok { return nil, true, errors.New(errorIncorrectFormat); }

	removeErr := user.RemoveFriend(friendName);
	if(removeErr != nil){ return nil, true, removeErr; }

	//
	return friendName, true, nil;
}
