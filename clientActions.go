package gopher

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/actions"
	"github.com/hewiefreeman/GopherGameServer/database"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"github.com/hewiefreeman/GopherGameServer/users"
	"sync"
)

const (
	errorInvalidAction               = "Invalid action"
	errorLoggedIn                    = "You must be logged out"
	errorNotLoggedIn                 = "You must be logged in"
	errorFeatureDisabled             = "Server feature is disabled"
	errorRoomControl                 = "Clients cannot control rooms"
	errorServerRoom                  = "Clients cannot control that room type"
	errorNotOwner                    = "You are not the owner of the room"
	errorIncorrectFormat             = "Incorrect data format"
	errorIncorrectFormatName         = "Incorrect data format for user name"
	errorIncorrectFormatPass         = "Incorrect data format for password"
	errorIncorrectFormatNewPass      = "Incorrect data format for new password"
	errorIncorrectFormatAction       = "Incorrect data format for action"
	errorIncorrectFormatCols         = "Incorrect data format for custom columns"
	errorIncorrectFormatRemember     = "Incorrect data format for remember me"
	errorIncorrectFormatGuest        = "Incorrect data format for guest"
	errorIncorrectFormatRoomName     = "Incorrect data format for room name"
	errorIncorrectFormatRoomType     = "Incorrect data format for room type"
	errorIncorrectFormatPrivateRoom  = "Incorrect data format for private room"
	errorIncorrectFormatMaxRoomUsers = "Incorrect data format for max room users"
	errorIncorrectFormatVarKey       = "Incorrect data format for variable key"
)

func clientActionHandler(action clientAction, user **users.User, conn *websocket.Conn,
	deviceTag *string, devicePass *string, deviceUserID *int, connID *string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	switch _action := action.A; _action {

	// HIGH LOOK-UP PRIORITY ITEMS

	case helpers.ClientActionCustomAction:
		return clientCustomAction(action.P, user, conn, *connID, clientMux)
	case helpers.ClientActionVoiceStream:
		return clientActionVoiceStream(action.P, user, conn, *connID, clientMux)

	// USER VARIABLES

	case helpers.ClientActionSetVariable:
		return clientActionSetVariable(action.P, user, *connID, clientMux)
	case helpers.ClientActionSetVariables:
		return clientActionSetVariables(action.P, user, *connID, clientMux)

	// CHAT

	case helpers.ClientActionChatMessage:
		return clientActionChatMessage(action.P, user, *connID, clientMux)
	case helpers.ClientActionPrivateMessage:
		return clientActionPrivateMessage(action.P, user, *connID, clientMux)

	// CHANGE STATUS

	case helpers.ClientActionChangeStatus:
		return clientActionChangeStatus(action.P, user, clientMux)

	// LOGIN/LOGOUT

	case helpers.ClientActionLogin:
		return clientActionLogin(action.P, user, deviceTag, devicePass, deviceUserID, conn, connID, clientMux)
	case helpers.ClientActionLogout:
		return clientActionLogout(user, deviceTag, devicePass, deviceUserID, connID, clientMux)

	// ROOM ACTIONS

	case helpers.ClientActionJoinRoom:
		return clientActionJoinRoom(action.P, user, *connID, clientMux)
	case helpers.ClientActionLeaveRoom:
		return clientActionLeaveRoom(user, *connID, clientMux)
	case helpers.ClientActionCreateRoom:
		return clientActionCreateRoom(action.P, user, *connID, clientMux)
	case helpers.ClientActionDeleteRoom:
		return clientActionDeleteRoom(action.P, user, clientMux)
	case helpers.ClientActionRoomInvite:
		return clientActionRoomInvite(action.P, user, *connID, clientMux)
	case helpers.ClientActionRevokeInvite:
		return clientActionRevokeInvite(action.P, user, *connID, clientMux)

	// FRIENDING

	case helpers.ClientActionFriendRequest:
		return clientActionFriendRequest(action.P, user, clientMux)
	case helpers.ClientActionAcceptFriend:
		return clientActionAcceptFriend(action.P, user, clientMux)
	case helpers.ClientActionDeclineFriend:
		return clientActionDeclineFriend(action.P, user, clientMux)
	case helpers.ClientActionRemoveFriend:
		return clientActionRemoveFriend(action.P, user, clientMux)

	// DATABASE

	case helpers.ClientActionSignup:
		return clientActionSignup(action.P, user, clientMux)
	case helpers.ClientActionDeleteAccount:
		return clientActionDeleteAccount(action.P, user, clientMux)
	case helpers.ClientActionChangePassword:
		return clientActionChangePassword(action.P, user, clientMux)
	case helpers.ClientActionChangeAccountInfo:
		return clientActionChangeAccountInfo(action.P, user, clientMux)

	// INVALID CLIENT ACTION

	default:
		return nil, true, helpers.NewError(errorInvalidAction, helpers.Error_Gopher_Invalid_Action)
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CUSTOM CLIENT ACTIONS   /////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientCustomAction(params interface{}, user **users.User, conn *websocket.Conn, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	var ok bool
	var pMap map[string]interface{}
	var action string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	if action, ok = pMap["a"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatAction, helpers.Error_Gopher_Incorrect_Custom_Action)
	}
	(*clientMux).Lock()
	userRef := *user
	(*clientMux).Unlock()
	actions.HandleCustomClientAction(action, pMap["d"], userRef, conn, connID)
	return nil, false, helpers.NewError("", 0)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHANGE USER STATUS   ////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionChangeStatus(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var statusF float64
	if statusF, ok = params.(float64); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	status := int(statusF)
	//
	userRef.SetStatus(status)
	//
	return status, true, helpers.NewError("", 0)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ACCOUNT/DATABASE ACTIONS   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionSignup(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user != nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorLoggedIn, helpers.Error_Gopher_Logged_In)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.Error_Gopher_Feature_Disabled)
	}
	(*clientMux).Unlock()
	//GET ITEMS FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var customCols map[string]interface{}
	var userName string
	var pass string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatCols, helpers.Error_Gopher_Columns_Format)
		}
	}
	if userName, ok = pMap["n"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatName, helpers.Error_Gopher_Name_Format)
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatPass, helpers.Error_Gopher_Password_Format)
	}
	//SIGN CLIENT UP
	signupErr := database.SignUpClient(userName, pass, customCols)
	if signupErr != 0 {
		return nil, true, helpers.NewError(signupErr.Error(), helpers.Error_Gopher_Sign_Up)
	}

	//
	return nil, true, helpers.NewError("", 0)
}

func clientActionDeleteAccount(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user != nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorLoggedIn, helpers.Error_Gopher_Logged_In)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.Error_Gopher_Feature_Disabled)
	}
	(*clientMux).Unlock()
	//GET ITEMS FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var customCols map[string]interface{}
	var userName string
	var pass string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatCols, helpers.Error_Gopher_Columns_Format)
		}
	}
	if userName, ok = pMap["n"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatName, helpers.Error_Gopher_Name_Format)
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatPass, helpers.Error_Gopher_Password_Format)
	}

	//CHECK IF USER IS ONLINE
	_, err := users.Get(userName)
	if err == nil {
		return nil, true, helpers.NewError(errorLoggedIn.Error(), helpers.Error_Gopher_Logged_In)
	}

	//DELETE ACCOUNT
	deleteErr := database.DeleteAccount(userName, pass, customCols)
	if deleteErr != nil {
		return nil, true, helpers.NewError(deleteErr.Error(), helpers.Error_Gopher_Delete_Account_Error)
	}
	//
	return nil, true, helpers.NewError("", 0)
}

func clientActionChangePassword(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.Error_Gopher_Feature_Disabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET ITEMS FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var customCols map[string]interface{}
	var pass string
	var newPass string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Columns_Format)
		}
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatPass, helpers.Error_Gopher_Password_Format)
	}
	if newPass, ok = pMap["n"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatNewPass, helpers.Error_Gopher_New_Password_Format)
	}
	//CHANGE PASSWORD
	changeErr := database.ChangePassword(userRef.Name(), pass, newPass, customCols)
	if changeErr != nil {
		return nil, true, helpers.NewError(changeErr.Error(), helpers.Error_Gopher_Password_Change)
	}

	//
	return nil, true, helpers.NewError("", 0)
}

func clientActionChangeAccountInfo(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.Error_Gopher_Feature_Disabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET ITEMS FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var customCols map[string]interface{}
	var pass string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatCols, helpers.Error_Gopher_Columns_Format)
		}
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatPass, helpers.Error_Gopher_Password_Format)
	}
	//CHANGE ACCOUNT INFO
	changeErr := database.ChangeAccountInfo(userRef.Name(), pass, customCols)
	if changeErr != nil {
		return nil, true, helpers.NewError(changeErr.Error(), helpers.Error_Gopher_Info_Change)
	}

	//
	return nil, true, helpers.NewError("", 0)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOGIN+LOGOUT ACTIONS   //////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionLogin(params interface{}, user **users.User, deviceTag *string, devicePass *string, deviceUserID *int, conn *websocket.Conn,
	connID *string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user != nil {
		defer (*clientMux).Unlock()
		return nil, true, helpers.NewError(errorLoggedIn, helpers.Error_Gopher_Logged_In)
	}
	(*clientMux).Unlock()
	//MAKE A MAP FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var name string
	var pass string
	var remMe bool = false
	var guest bool = false
	var customCols map[string]interface{}
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	if name, ok = pMap["n"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatName, helpers.Error_Gopher_Name_Format)
	}
	if (*settings).EnableSqlFeatures {
		if pass, ok = pMap["p"].(string); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatPass, helpers.Error_Gopher_Password_Format)
		}
		if (*settings).RememberMe {
			if remMe, ok = pMap["r"].(bool); !ok {
				return nil, true, helpers.NewError(errorIncorrectFormatRemember, helpers.Error_Gopher_Remember_Format)
			}
		}
	}
	if pMap["g"] != nil {
		if guest, ok = pMap["g"].(bool); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatGuest, helpers.Error_Gopher_Guest_Format)
		}
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatCols, helpers.Error_Gopher_Columns_Format)
		}
	}
	//LOG IN
	var dbIndex int
	var uName string
	var dPass string
	var cID string
	var err error
	if (*settings).EnableSqlFeatures {
		uName, dbIndex, dPass, err = database.LoginClient(name, pass, *deviceTag, remMe, customCols)
		if err != nil {
			return nil, true, helpers.NewError(err.Error(), helpers.Error_Gopher_Login)
		}
		cID, err = users.Login(uName, dbIndex, dPass, guest, remMe, conn, user, clientMux)
	} else {
		cID, err = users.Login(name, -1, "", guest, false, conn, user, clientMux)
	}
	if err != nil {
		return nil, true, helpers.NewError(err.Error(), helpers.Error_Gopher_Login)
	}
	//CHANGE SOCKET'S userName
	*devicePass = dPass
	*deviceUserID = dbIndex
	*connID = cID

	//
	return nil, false, helpers.NewError("", 0)
}

func clientActionLogout(user **users.User, deviceTag *string, devicePass *string, deviceUserID *int, connID *string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	}
	userRef := *user
	(*clientMux).Unlock()
	//LOG User OUT AND RESET
	userRef.Logout(*connID)
	//REMOVE AUTO-LOG IF ANY
	if (*settings).EnableSqlFeatures && (*settings).RememberMe {
		database.RemoveAutoLog(*deviceUserID, *deviceTag)
	}
	//
	*devicePass = ""
	*deviceUserID = 0
	*connID = ""

	//
	return nil, false, helpers.NewError("", 0)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ROOM ACTIONS   //////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionJoinRoom(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET ROOM NAME FROM PARAMS
	var ok bool
	var roomName string
	if roomName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	//GET ROOM
	room, roomErr := rooms.Get(roomName)
	if roomErr != nil {
		return nil, true, helpers.NewError(roomErr.Error(), helpers.Error_Gopher_Join)
	}
	//MAKE User JOIN THE Room
	joinErr := userRef.Join(room, connID)
	if joinErr != nil {
		return nil, true, helpers.NewError(joinErr.Error(), helpers.Error_Gopher_Join)
	}

	//
	return nil, true, helpers.NewError("", 0)
}

func clientActionLeaveRoom(user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	}
	userRef := *user
	(*clientMux).Unlock()
	//MAKE USER LEAVE ROOM
	leaveErr := userRef.Leave(connID)
	if leaveErr != nil {
		return nil, true, helpers.NewError(leaveErr.Error(), helpers.Error_Gopher_Leave)
	}

	//
	return nil, true, helpers.NewError("", 0)
}

func clientActionCreateRoom(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorRoomControl, helpers.Error_Gopher_Room_Control)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var pMap map[string]interface{}
	var roomName string
	var roomType string
	var private bool
	var maxUsersF float64
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	if roomName, ok = pMap["n"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatRoomName, helpers.Error_Gopher_Room_Name_Format)
	}
	if roomType, ok = pMap["t"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatRoomType, helpers.Error_Gopher_Room_Type_Format)
	}
	if private, ok = pMap["t"].(bool); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatPrivateRoom, helpers.Error_Gopher_Private_Format)
	}
	if maxUsersF, ok = pMap["m"].(float64); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatMaxRoomUsers, helpers.Error_Gopher_Max_Room_Format)
	}
	maxUsers := int(maxUsersF)
	//
	if rType, ok := rooms.GetRoomTypes()[roomType]; !ok {
		return nil, true, helpers.NewError(errorRoomType, helpers.Error_Gopher_Max_Room_Format)
	} else if rType.ServerOnly() {
		return nil, true, helpers.NewError(errorServerRoom, helpers.Error_Gopher_Server_Room)
	}
	//CHECK IF USER IS IN A ROOM
	/*currRoom := user.RoomIn()
	if currRoom != nil && currRoom.Name() != "" {
		return nil, true, errors.New("You must leave your current room to create a room")
	}*/
	//MAKE THE Room
	room, roomErr := rooms.New(roomName, roomType, private, maxUsers, userRef.Name())
	if roomErr != nil {
		return nil, true, helpers.NewError(roomErr.Error(), helpers.Error_Gopher_Create_Room)
	}
	//ADD THE User TO THE ROOM
	joinErr := userRef.Join(room, connID)
	if joinErr != nil {
		return nil, true, helpers.NewError(joinErr.Error(), helpers.Error_Gopher_Join)
	}

	//
	return roomName, true, helpers.NewError("", 0)
}

func clientActionDeleteRoom(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorRoomControl, helpers.Error_Gopher_Room_Control)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var roomName string
	if roomName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	//GET ROOM
	room, roomErr := rooms.Get(roomName)
	if roomErr != nil {
		return nil, true, helpers.NewError(roomErr.Error(), helpers.Error_Gopher_Delete_Room)
	} else if room.Owner() != userRef.Name() {
		return nil, true, helpers.NewError(errorNotOwner, helpers.Error_Gopher_Not_Owner)
	}
	//
	rType := rooms.GetRoomTypes()[room.Type()]
	if rType.ServerOnly() {
		return nil, true, helpers.NewError(errorServerRoom, helpers.Error_Gopher_Server_Room)
	}
	//DELETE ROOM
	deleteErr := room.Delete()
	if deleteErr != nil {
		return nil, true, helpers.NewError(deleteErr.Error(), helpers.Error_Gopher_Delete_Room)
	}

	return nil, true, helpers.NewError("", 0)
}

func clientActionRoomInvite(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorRoomControl, helpers.Error_Gopher_Room_Control)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var name string
	if name, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	//GET INVITED USER
	invUser, invUserErr := users.Get(name)
	if invUserErr != nil {
		return nil, true, helpers.NewError(invUserErr.Error(), helpers.Error_Gopher_Invite)
	}
	//INVITE
	invUserErr = userRef.Invite(invUser, connID)
	if invUserErr != nil {
		return nil, true, helpers.NewError(invUserErr.Error(), helpers.Error_Gopher_Invite)
	}
	//
	return nil, true, helpers.NewError("", 0)
}

func clientActionRevokeInvite(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorRoomControl, helpers.Error_Gopher_Room_Control)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var name string
	if name, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}
	//REVOKE INVITE
	revokeErr := userRef.RevokeInvite(name, connID)
	if revokeErr != nil {
		return nil, true, helpers.NewError(revokeErr.Error, helpers.Error_Gopher_Revoke_Invite)
	}
	//
	return nil, true, helpers.NewError("", 0)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHAT+VOICE ACTIONS   ////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionVoiceStream(params interface{}, user **users.User, conn *websocket.Conn, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, helpers.NewError("", 0)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET CURRENT ROOM
	currRoom := userRef.RoomIn(connID)
	if currRoom == nil || currRoom.Name() == "" {
		return nil, false, helpers.NewError("", 0)
	}
	//CHECK IF VOICE CHAT ROOM
	rType := rooms.GetRoomTypes()[currRoom.Type()]
	if !rType.VoiceChatEnabled() {
		return nil, false, helpers.NewError("", 0)
	}
	//SEND VOICE STREAM
	currRoom.VoiceStream(userRef.Name(), conn, params)
	//
	return nil, false, helpers.NewError("", 0)
}

func clientActionChatMessage(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, helpers.NewError("", 0)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET CURRENT ROOM
	currRoom := userRef.RoomIn(connID)
	if currRoom == nil || currRoom.Name() == "" {
		return nil, false, helpers.NewError("", 0)
	}
	//SEND CHAT MESSAGE
	currRoom.ChatMessage(userRef.Name(), params)
	//
	return nil, false, helpers.NewError("", 0)
}

func clientActionPrivateMessage(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, helpers.NewError("", 0)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var pMap map[string]interface{}
	var userName string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, false, helpers.NewError("", 0)
	}
	if userName, ok = pMap["u"].(string); !ok {
		return nil, false, helpers.NewError("", 0)
	}
	//GET CURRENT ROOM
	currRoom := userRef.RoomIn(connID)
	if currRoom == nil || currRoom.Name() == "" {
		return nil, false, helpers.NewError("", 0)
	}
	//SEND CHAT MESSAGE
	userRef.PrivateMessage(userName, pMap["m"])
	//
	return nil, false, helpers.NewError("", 0)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   ////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionSetVariable(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, helpers.NewError("", 0)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var pMap map[string]interface{}
	var varKey string
	var varVal interface{}
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, false, helpers.NewError("", 0)
	}
	if varKey, ok = pMap["k"].(string); !ok {
		return nil, false, helpers.NewError("", 0)
	}
	varVal = pMap["v"]
	//SET THE VARIABLE
	userRef.SetVariable(varKey, varVal, connID)
	//
	return nil, false, helpers.NewError("", 0)
}

func clientActionSetVariables(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, helpers.NewError("", 0)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var pMap map[string]interface{}
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, false, helpers.NewError("", 0)
	}
	//SET THE VARIABLES
	userRef.SetVariables(pMap, connID)
	//
	return nil, false, helpers.NewError("", 0)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   FRIENDING ACTIONS   /////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionFriendRequest(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.Error_Gopher_Feature_Disabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}

	requestErr := userRef.FriendRequest(friendName)
	if requestErr != nil {
		return nil, true, helpers.NewError(requestErr.Error(), helpers.Error_Gopher_Friend_Request)
	}

	//
	return nil, false, helpers.NewError("", 0)
}

func clientActionAcceptFriend(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.Error_Gopher_Feature_Disabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}

	acceptErr := userRef.AcceptFriendRequest(friendName)
	if acceptErr != nil {
		return nil, true, helpers.NewError(acceptErr.Error(), helpers.Error_Gopher_Friend_Accept)
	}

	//
	return nil, false, helpers.NewError("", 0)
}

func clientActionDeclineFriend(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.Error_Gopher_Feature_Disabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}

	declineErr := userRef.DeclineFriendRequest(friendName)
	if declineErr != nil {
		return nil, true, helpers.NewError(declineErr.Error(), helpers.Error_Gopher_Friend_Decline)
	}

	//
	return nil, false, helpers.NewError("", 0)
}

func clientActionRemoveFriend(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.Error_Gopher_Not_Logged_In)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.Error_Gopher_Feature_Disabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.Error_Gopher_Incorrect_Format)
	}

	removeErr := userRef.RemoveFriend(friendName)
	if removeErr != nil {
		return nil, true, helpers.NewError(removeErr.Error(), helpers.Error_Gopher_Friend_Remove)
	}

	//
	return nil, false, helpers.NewError("", 0)
}
