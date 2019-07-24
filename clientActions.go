package gopher

import (
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/actions"
	"github.com/hewiefreeman/GopherGameServer/core"
	"github.com/hewiefreeman/GopherGameServer/database"
	"github.com/hewiefreeman/GopherGameServer/helpers"
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
	errorRoomType                    = "Invalid room type"
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

func clientActionHandler(action clientAction, user **core.User, conn *websocket.Conn,
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
		return nil, true, helpers.NewError(errorInvalidAction, helpers.ErrorGopherInvalidAction)
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CUSTOM CLIENT ACTIONS   /////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientCustomAction(params interface{}, user **core.User, conn *websocket.Conn, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	var ok bool
	var pMap map[string]interface{}
	var action string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	if action, ok = pMap["a"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatAction, helpers.ErrorGopherIncorrectCustomAction)
	}
	(*clientMux).Lock()
	userRef := *user
	(*clientMux).Unlock()
	actions.HandleCustomClientAction(action, pMap["d"], userRef, conn, connID)
	return nil, false, helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHANGE USER STATUS   ////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionChangeStatus(params interface{}, user **core.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorLoggedIn, helpers.ErrorGopherNotLoggedIn)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var statusF float64
	if statusF, ok = params.(float64); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	status := int(statusF)
	//
	userRef.SetStatus(status)
	//
	return status, true, helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ACCOUNT/DATABASE ACTIONS   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionSignup(params interface{}, user **core.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user != nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorLoggedIn, helpers.ErrorGopherLoggedIn)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.ErrorGopherFeatureDisabled)
	}
	(*clientMux).Unlock()
	//GET ITEMS FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var customCols map[string]interface{}
	var userName string
	var pass string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatCols, helpers.ErrorGopherColumnsFormat)
		}
	}
	if userName, ok = pMap["n"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatName, helpers.ErrorGopherNameFormat)
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatPass, helpers.ErrorGopherPasswordFormat)
	}
	//SIGN CLIENT UP
	signupErr := database.SignUpClient(userName, pass, customCols)
	if signupErr.ID != 0 {
		return nil, true, signupErr
	}

	//
	return nil, true, helpers.NoError()
}

func clientActionDeleteAccount(params interface{}, user **core.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user != nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorLoggedIn, helpers.ErrorGopherLoggedIn)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.ErrorGopherFeatureDisabled)
	}
	(*clientMux).Unlock()
	//GET ITEMS FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var customCols map[string]interface{}
	var userName string
	var pass string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatCols, helpers.ErrorGopherColumnsFormat)
		}
	}
	if userName, ok = pMap["n"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatName, helpers.ErrorGopherNameFormat)
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatPass, helpers.ErrorGopherPasswordFormat)
	}

	//CHECK IF USER IS ONLINE
	_, err := core.GetUser(userName)
	if err == nil {
		return nil, true, helpers.NewError(err.Error(), helpers.ErrorGopherLoggedIn)
	}

	//DELETE ACCOUNT
	deleteErr := database.DeleteAccount(userName, pass, customCols)
	if deleteErr.ID != 0 {
		return nil, true, deleteErr
	}
	//
	return nil, true, helpers.NoError()
}

func clientActionChangePassword(params interface{}, user **core.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.ErrorGopherFeatureDisabled)
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
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherColumnsFormat)
		}
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatPass, helpers.ErrorGopherPasswordFormat)
	}
	if newPass, ok = pMap["n"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatNewPass, helpers.ErrorGopherNewPasswordFormat)
	}
	//CHANGE PASSWORD
	changeErr := database.ChangePassword(userRef.Name(), pass, newPass, customCols)
	if changeErr.ID != 0 {
		return nil, true, changeErr
	}

	//
	return nil, true, helpers.NoError()
}

func clientActionChangeAccountInfo(params interface{}, user **core.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.ErrorGopherFeatureDisabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET ITEMS FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var customCols map[string]interface{}
	var pass string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatCols, helpers.ErrorGopherColumnsFormat)
		}
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatPass, helpers.ErrorGopherPasswordFormat)
	}
	//CHANGE ACCOUNT INFO
	changeErr := database.ChangeAccountInfo(userRef.Name(), pass, customCols)
	if changeErr.ID != 0 {
		return nil, true, changeErr
	}

	//
	return nil, true, helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOGIN+LOGOUT ACTIONS   //////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionLogin(params interface{}, user **core.User, deviceTag *string, devicePass *string, deviceUserID *int, conn *websocket.Conn,
	connID *string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user != nil {
		defer (*clientMux).Unlock()
		return nil, true, helpers.NewError(errorLoggedIn, helpers.ErrorGopherLoggedIn)
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
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	if name, ok = pMap["n"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatName, helpers.ErrorGopherNameFormat)
	}
	if (*settings).EnableSqlFeatures {
		if pass, ok = pMap["p"].(string); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatPass, helpers.ErrorGopherPasswordFormat)
		}
		if (*settings).RememberMe {
			if remMe, ok = pMap["r"].(bool); !ok {
				return nil, true, helpers.NewError(errorIncorrectFormatRemember, helpers.ErrorGopherRememberFormat)
			}
		}
	}
	if pMap["g"] != nil {
		if guest, ok = pMap["g"].(bool); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatGuest, helpers.ErrorGopherGuestFormat)
		}
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, helpers.NewError(errorIncorrectFormatCols, helpers.ErrorGopherColumnsFormat)
		}
	}
	//LOG IN
	var dbIndex int
	var uName string
	var dPass string
	var cID string
	var err helpers.GopherError
	if (*settings).EnableSqlFeatures && !guest {
		uName, dbIndex, dPass, err = database.LoginClient(name, pass, *deviceTag, remMe, customCols)
		if err.ID != 0 {
			return nil, true, err
		}
		cID, err = core.Login(uName, dbIndex, dPass, guest, remMe, conn, user, clientMux)
	} else {
		cID, err = core.Login(name, -1, "", guest, false, conn, user, clientMux)
	}
	if err.ID != 0 {
		return nil, true, err
	}
	//CHANGE SOCKET'S userName
	*devicePass = dPass
	*deviceUserID = dbIndex
	*connID = cID

	//
	return nil, false, helpers.NoError()
}

func clientActionLogout(user **core.User, deviceTag *string, devicePass *string, deviceUserID *int, connID *string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
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
	return nil, false, helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ROOM ACTIONS   //////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionJoinRoom(params interface{}, user **core.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET ROOM NAME FROM PARAMS
	var ok bool
	var roomName string
	if roomName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	//GET ROOM
	room, roomErr := core.GetRoom(roomName)
	if roomErr != nil {
		return nil, true, helpers.NewError(roomErr.Error(), helpers.ErrorGopherJoin)
	}
	//MAKE User JOIN THE Room
	joinErr := userRef.Join(room, connID)
	if joinErr != nil {
		return nil, true, helpers.NewError(joinErr.Error(), helpers.ErrorGopherJoin)
	}

	//
	return nil, false, helpers.NoError()
}

func clientActionLeaveRoom(user **core.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	}
	userRef := *user
	(*clientMux).Unlock()
	//MAKE USER LEAVE ROOM
	leaveErr := userRef.Leave(connID)
	if leaveErr != nil {
		return nil, true, helpers.NewError(leaveErr.Error(), helpers.ErrorGopherLeave)
	}

	//
	return nil, false, helpers.NoError()
}

func clientActionCreateRoom(params interface{}, user **core.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorRoomControl, helpers.ErrorGopherRoomControl)
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
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	if roomName, ok = pMap["n"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatRoomName, helpers.ErrorGopherRoomNameFormat)
	}
	if roomType, ok = pMap["t"].(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatRoomType, helpers.ErrorGopherRoomTypeFormat)
	}
	if private, ok = pMap["t"].(bool); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatPrivateRoom, helpers.ErrorGopherPrivateFormat)
	}
	if maxUsersF, ok = pMap["m"].(float64); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormatMaxRoomUsers, helpers.ErrorGopherMaxRoomFormat)
	}
	maxUsers := int(maxUsersF)
	//
	if rType, ok := core.GetRoomTypes()[roomType]; !ok {
		return nil, true, helpers.NewError(errorRoomType, helpers.ErrorGopherMaxRoomFormat)
	} else if rType.ServerOnly() {
		return nil, true, helpers.NewError(errorServerRoom, helpers.ErrorGopherServerRoom)
	}
	//CHECK IF USER IS IN A ROOM
	/*currRoom := user.RoomIn()
	if currRoom != nil && currRoom.Name() != "" {
		return nil, true, errors.New("You must leave your current room to create a room")
	}*/
	//MAKE THE Room
	room, roomErr := core.NewRoom(roomName, roomType, private, maxUsers, userRef.Name())
	if roomErr != nil {
		return nil, true, helpers.NewError(roomErr.Error(), helpers.ErrorGopherCreateRoom)
	}
	//ADD THE User TO THE ROOM
	joinErr := userRef.Join(room, connID)
	if joinErr != nil {
		return nil, true, helpers.NewError(joinErr.Error(), helpers.ErrorGopherJoin)
	}

	//
	return roomName, true, helpers.NoError()
}

func clientActionDeleteRoom(params interface{}, user **core.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorRoomControl, helpers.ErrorGopherRoomControl)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var roomName string
	if roomName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	//GET ROOM
	room, roomErr := core.GetRoom(roomName)
	if roomErr != nil {
		return nil, true, helpers.NewError(roomErr.Error(), helpers.ErrorGopherDeleteRoom)
	} else if room.Owner() != userRef.Name() {
		return nil, true, helpers.NewError(errorNotOwner, helpers.ErrorGopherNotOwner)
	}
	//
	rType := core.GetRoomTypes()[room.Type()]
	if rType.ServerOnly() {
		return nil, true, helpers.NewError(errorServerRoom, helpers.ErrorGopherServerRoom)
	}
	//DELETE ROOM
	deleteErr := room.Delete()
	if deleteErr != nil {
		return nil, true, helpers.NewError(deleteErr.Error(), helpers.ErrorGopherDeleteRoom)
	}

	return roomName, true, helpers.NoError()
}

func clientActionRoomInvite(params interface{}, user **core.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorRoomControl, helpers.ErrorGopherRoomControl)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var name string
	if name, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	//GET INVITED USER
	invUser, invUserErr := core.GetUser(name)
	if invUserErr != nil {
		return nil, true, helpers.NewError(invUserErr.Error(), helpers.ErrorGopherInvite)
	}
	//INVITE
	invUserErr = userRef.Invite(invUser, connID)
	if invUserErr != nil {
		return nil, true, helpers.NewError(invUserErr.Error(), helpers.ErrorGopherInvite)
	}
	//
	return nil, true, helpers.NoError()
}

func clientActionRevokeInvite(params interface{}, user **core.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorRoomControl, helpers.ErrorGopherRoomControl)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var name string
	if name, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}
	//REVOKE INVITE
	revokeErr := userRef.RevokeInvite(name, connID)
	if revokeErr != nil {
		return nil, true, helpers.NewError(revokeErr.Error(), helpers.ErrorGopherRevokeInvite)
	}
	//
	return nil, true, helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHAT+VOICE ACTIONS   ////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionVoiceStream(params interface{}, user **core.User, conn *websocket.Conn, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, helpers.NoError()
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET CURRENT ROOM
	currRoom := userRef.RoomIn(connID)
	if currRoom == nil || currRoom.Name() == "" {
		return nil, false, helpers.NoError()
	}
	//CHECK IF VOICE CHAT ROOM
	rType := core.GetRoomTypes()[currRoom.Type()]
	if !rType.VoiceChatEnabled() {
		return nil, false, helpers.NoError()
	}
	//SEND VOICE STREAM
	currRoom.VoiceStream(userRef.Name(), conn, params)
	//
	return nil, false, helpers.NoError()
}

func clientActionChatMessage(params interface{}, user **core.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, helpers.NoError()
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET CURRENT ROOM
	currRoom := userRef.RoomIn(connID)
	if currRoom == nil || currRoom.Name() == "" {
		return nil, false, helpers.NoError()
	}
	//SEND CHAT MESSAGE
	currRoom.ChatMessage(userRef.Name(), params)
	//
	return nil, false, helpers.NoError()
}

func clientActionPrivateMessage(params interface{}, user **core.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, helpers.NoError()
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var pMap map[string]interface{}
	var userName string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, false, helpers.NoError()
	}
	if userName, ok = pMap["u"].(string); !ok {
		return nil, false, helpers.NoError()
	}
	//GET CURRENT ROOM
	currRoom := userRef.RoomIn(connID)
	if currRoom == nil || currRoom.Name() == "" {
		return nil, false, helpers.NoError()
	}
	//SEND CHAT MESSAGE
	userRef.PrivateMessage(userName, pMap["m"])
	//
	return nil, false, helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   ////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionSetVariable(params interface{}, user **core.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, helpers.NoError()
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var pMap map[string]interface{}
	var varKey string
	var varVal interface{}
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, false, helpers.NoError()
	}
	if varKey, ok = pMap["k"].(string); !ok {
		return nil, false, helpers.NoError()
	}
	varVal = pMap["v"]
	//SET THE VARIABLE
	userRef.SetVariable(varKey, varVal, connID)
	//
	return nil, false, helpers.NoError()
}

func clientActionSetVariables(params interface{}, user **core.User, connID string, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, helpers.NoError()
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var pMap map[string]interface{}
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, false, helpers.NoError()
	}
	//SET THE VARIABLES
	userRef.SetVariables(pMap, connID)
	//
	return nil, false, helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   FRIENDING ACTIONS   /////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionFriendRequest(params interface{}, user **core.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.ErrorGopherFeatureDisabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}

	requestErr := userRef.FriendRequest(friendName)
	if requestErr != nil {
		return nil, true, helpers.NewError(requestErr.Error(), helpers.ErrorGopherFriendRequest)
	}

	//
	return nil, false, helpers.NoError()
}

func clientActionAcceptFriend(params interface{}, user **core.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.ErrorGopherFeatureDisabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}

	acceptErr := userRef.AcceptFriendRequest(friendName)
	if acceptErr != nil {
		return nil, true, helpers.NewError(acceptErr.Error(), helpers.ErrorGopherFriendAccept)
	}

	//
	return nil, false, helpers.NoError()
}

func clientActionDeclineFriend(params interface{}, user **core.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.ErrorGopherFeatureDisabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}

	declineErr := userRef.DeclineFriendRequest(friendName)
	if declineErr != nil {
		return nil, true, helpers.NewError(declineErr.Error(), helpers.ErrorGopherFriendDecline)
	}

	//
	return nil, false, helpers.NoError()
}

func clientActionRemoveFriend(params interface{}, user **core.User, clientMux *sync.Mutex) (interface{}, bool, helpers.GopherError) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorNotLoggedIn, helpers.ErrorGopherNotLoggedIn)
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, helpers.NewError(errorFeatureDisabled, helpers.ErrorGopherFeatureDisabled)
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, helpers.NewError(errorIncorrectFormat, helpers.ErrorGopherIncorrectFormat)
	}

	removeErr := userRef.RemoveFriend(friendName)
	if removeErr != nil {
		return nil, true, helpers.NewError(removeErr.Error(), helpers.ErrorGopherFriendRemove)
	}

	//
	return nil, false, helpers.NoError()
}
