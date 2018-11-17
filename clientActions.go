package gopher

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/actions"
	"github.com/hewiefreeman/GopherGameServer/database"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"github.com/hewiefreeman/GopherGameServer/users"
	"github.com/mssola/user_agent"
	"sync"
)

const (
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

func clientActionHandler(action clientAction, user **users.User, conn *websocket.Conn, ua *user_agent.UserAgent,
	deviceTag *string, devicePass *string, deviceUserID *int, connID *string, clientMux *sync.Mutex) (interface{}, bool, error) {
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
		return nil, true, errors.New("Unrecognized client action")
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CUSTOM CLIENT ACTIONS   /////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientCustomAction(params interface{}, user **users.User, conn *websocket.Conn, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	var ok bool
	var pMap map[string]interface{}
	var action string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}
	if action, ok = pMap["a"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatAction)
	}
	(*clientMux).Lock()
	userRef := *user
	(*clientMux).Unlock()
	actions.HandleCustomClientAction(action, pMap["d"], userRef, conn, connID)
	return nil, false, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHANGE USER STATUS   ////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionChangeStatus(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to change your status")
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var statusF float64
	if statusF, ok = params.(float64); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}
	status := int(statusF)
	//
	statusErr := userRef.SetStatus(status)
	if statusErr != nil {
		return nil, true, statusErr
	}
	//
	return status, true, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ACCOUNT/DATABASE ACTIONS   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionSignup(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user != nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged out to sign up")
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, errors.New("Required server features are not enabled")
	}
	(*clientMux).Unlock()
	//GET ITEMS FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var customCols map[string]interface{}
	var userName string
	var pass string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, errors.New(errorIncorrectFormatCols)
		}
	}
	if userName, ok = pMap["n"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatName)
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatPass)
	}
	//SIGN CLIENT UP
	signupErr := database.SignUpClient(userName, pass, customCols)
	if signupErr != nil {
		return nil, true, signupErr
	}

	//
	return nil, true, nil
}

func clientActionDeleteAccount(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user != nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged out to delete your account")
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, errors.New("Required server features are not enabled")
	}
	(*clientMux).Unlock()
	//GET ITEMS FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var customCols map[string]interface{}
	var userName string
	var pass string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, errors.New(errorIncorrectFormatCols)
		}
	}
	if userName, ok = pMap["n"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatName)
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatPass)
	}

	//CHECK IF USER IS ONLINE
	_, err := users.Get(userName)
	if err == nil {
		return nil, true, errors.New("The User must be logged off to delete their account.")
	}

	//DELETE ACCOUNT
	deleteErr := database.DeleteAccount(userName, pass, customCols)
	if deleteErr != nil {
		return nil, true, deleteErr
	}
	//
	return nil, true, nil
}

func clientActionChangePassword(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to change your password")
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, errors.New("Required server features are not enabled")
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
		return nil, true, errors.New(errorIncorrectFormat)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, errors.New(errorIncorrectFormatCols)
		}
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatPass)
	}
	if newPass, ok = pMap["n"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatNewPass)
	}
	//CHANGE PASSWORD
	changeErr := database.ChangePassword(userRef.Name(), pass, newPass, customCols)
	if changeErr != nil {
		return nil, true, changeErr
	}

	//
	return nil, true, nil
}

func clientActionChangeAccountInfo(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to change your account info")
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, errors.New("Required server features are not enabled")
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET ITEMS FROM PARAMS
	var ok bool
	var pMap map[string]interface{}
	var customCols map[string]interface{}
	var pass string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, errors.New(errorIncorrectFormatCols)
		}
	}
	if pass, ok = pMap["p"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatPass)
	}
	//CHANGE ACCOUNT INFO
	changeErr := database.ChangeAccountInfo(userRef.Name(), pass, customCols)
	if changeErr != nil {
		return nil, true, changeErr
	}

	//
	return nil, true, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOGIN+LOGOUT ACTIONS   //////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionLogin(params interface{}, user **users.User, deviceTag *string, devicePass *string, deviceUserID *int, conn *websocket.Conn,
	connID *string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user != nil {
		defer (*clientMux).Unlock()
		return nil, true, errors.New("Already logged in as '" + (*user).Name() + "'")
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
		return nil, true, errors.New(errorIncorrectFormat)
	}
	if name, ok = pMap["n"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatName)
	}
	if (*settings).EnableSqlFeatures {
		if pass, ok = pMap["p"].(string); !ok {
			return nil, true, errors.New(errorIncorrectFormatPass)
		}
		if (*settings).RememberMe {
			if remMe, ok = pMap["r"].(bool); !ok {
				return nil, true, errors.New(errorIncorrectFormatRemember)
			}
		}
	}
	if pMap["g"] != nil {
		if guest, ok = pMap["g"].(bool); !ok {
			return nil, true, errors.New(errorIncorrectFormatGuest)
		}
	}
	if pMap["c"] != nil {
		if customCols, ok = pMap["c"].(map[string]interface{}); !ok {
			return nil, true, errors.New(errorIncorrectFormatCols)
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
			return nil, true, err
		}
		cID, err = users.Login(uName, dbIndex, dPass, guest, remMe, conn, user, clientMux)
	} else {
		cID, err = users.Login(name, -1, "", guest, false, conn, user, clientMux)
	}
	if err != nil {
		return nil, true, err
	}
	//CHANGE SOCKET'S userName
	*devicePass = dPass
	*deviceUserID = dbIndex
	*connID = cID

	//
	return nil, false, nil
}

func clientActionLogout(user **users.User, deviceTag *string, devicePass *string, deviceUserID *int, connID *string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("Already logged out")
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
	return nil, false, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ROOM ACTIONS   //////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionJoinRoom(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to join a room")
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET ROOM NAME FROM PARAMS
	var ok bool
	var roomName string
	if roomName, ok = params.(string); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}
	//GET ROOM
	room, roomErr := rooms.Get(roomName)
	if roomErr != nil {
		return nil, true, roomErr
	}
	//MAKE User JOIN THE Room
	joinErr := userRef.Join(room, connID)
	if joinErr != nil {
		return nil, true, joinErr
	}

	//
	return nil, false, nil
}

func clientActionLeaveRoom(user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to leave a room")
	}
	userRef := *user
	(*clientMux).Unlock()
	//MAKE USER LEAVE ROOM
	leaveErr := userRef.Leave(connID)
	if leaveErr != nil {
		return nil, true, leaveErr
	}

	//
	return nil, false, nil
}

func clientActionCreateRoom(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to create a room")
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, errors.New("Clients do not have room control")
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
		return nil, true, errors.New(errorIncorrectFormat)
	}
	if roomName, ok = pMap["n"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatRoomName)
	}
	if roomType, ok = pMap["t"].(string); !ok {
		return nil, true, errors.New(errorIncorrectFormatRoomType)
	}
	if private, ok = pMap["t"].(bool); !ok {
		return nil, true, errors.New(errorIncorrectFormatPrivateRoom)
	}
	if maxUsersF, ok = pMap["m"].(float64); !ok {
		return nil, true, errors.New(errorIncorrectFormatMaxRoomUsers)
	}
	maxUsers := int(maxUsersF)
	//
	if rType, ok := rooms.GetRoomTypes()[roomType]; !ok {
		return nil, true, errors.New("Invalid room type")
	} else if rType.ServerOnly() {
		return nil, true, errors.New("Only the server can manipulate that type of room")
	}
	//CHECK IF USER IS IN A ROOM
	/*currRoom := user.RoomIn()
	if currRoom != nil && currRoom.Name() != "" {
		return nil, true, errors.New("You must leave your current room to create a room")
	}*/
	//MAKE THE Room
	room, roomErr := rooms.New(roomName, roomType, private, maxUsers, userRef.Name())
	if roomErr != nil {
		return nil, true, roomErr
	}
	//ADD THE User TO THE ROOM
	joinErr := userRef.Join(room, connID)
	if joinErr != nil {
		return nil, true, joinErr
	}

	//
	return roomName, true, nil
}

func clientActionDeleteRoom(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to delete a room")
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, errors.New("Clients do not have room control")
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var roomName string
	if roomName, ok = params.(string); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}
	//GET ROOM
	room, roomErr := rooms.Get(roomName)
	if roomErr != nil {
		return nil, true, roomErr
	} else if room.Owner() != userRef.Name() {
		return nil, true, errors.New("Only the owner of the room can delete it")
	}
	//
	rType := rooms.GetRoomTypes()[room.Type()]
	if rType.ServerOnly() {
		return nil, true, errors.New("Only the server can manipulate that type of room")
	}
	//DELETE ROOM
	deleteErr := room.Delete()
	if deleteErr != nil {
		return nil, true, deleteErr
	}

	return nil, true, nil
}

func clientActionRoomInvite(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to invite to a room")
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, errors.New("Clients do not have room control")
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var name string
	if name, ok = params.(string); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}
	//GET INVITED USER
	invUser, invUserErr := users.Get(name)
	if invUserErr != nil {
		return nil, true, invUserErr
	}
	//INVITE
	invUserErr = userRef.Invite(invUser, connID)
	if invUserErr != nil {
		return nil, true, invUserErr
	}
	//
	return nil, true, nil
}

func clientActionRevokeInvite(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to revoke an invite to a room")
	} else if !(*settings).UserRoomControl {
		(*clientMux).Unlock()
		return nil, true, errors.New("Clients do not have room control")
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var name string
	if name, ok = params.(string); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}
	//REVOKE INVITE
	revokeErr := userRef.RevokeInvite(name, connID)
	if revokeErr != nil {
		return nil, true, revokeErr
	}
	//
	return nil, true, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHAT+VOICE ACTIONS   ////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionVoiceStream(params interface{}, user **users.User, conn *websocket.Conn, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, nil
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET CURRENT ROOM
	currRoom := userRef.RoomIn(connID)
	if currRoom == nil || currRoom.Name() == "" {
		return nil, false, nil
	}
	//CHECK IF VOICE CHAT ROOM
	rType := rooms.GetRoomTypes()[currRoom.Type()]
	if !rType.VoiceChatEnabled() {
		return nil, false, nil
	}
	//SEND VOICE STREAM
	currRoom.VoiceStream(userRef.Name(), conn, params)
	//
	return nil, false, nil
}

func clientActionChatMessage(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, nil
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET CURRENT ROOM
	currRoom := userRef.RoomIn(connID)
	if currRoom == nil || currRoom.Name() == "" {
		return nil, false, nil
	}
	//SEND CHAT MESSAGE
	currRoom.ChatMessage(userRef.Name(), params)
	//
	return nil, false, nil
}

func clientActionPrivateMessage(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, nil
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var pMap map[string]interface{}
	var userName string
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, false, nil
	}
	if userName, ok = pMap["u"].(string); !ok {
		return nil, false, nil
	}
	//GET CURRENT ROOM
	currRoom := userRef.RoomIn(connID)
	if currRoom == nil || currRoom.Name() == "" {
		return nil, false, nil
	}
	//SEND CHAT MESSAGE
	userRef.PrivateMessage(userName, pMap["m"])
	//
	return nil, false, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   ////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionSetVariable(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, nil
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var pMap map[string]interface{}
	var varKey string
	var varVal interface{}
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, false, nil
	}
	if varKey, ok = pMap["k"].(string); !ok {
		return nil, false, nil
	}
	varVal = pMap["v"]
	//SET THE VARIABLE
	userRef.SetVariable(varKey, varVal, connID)
	//
	return nil, false, nil
}

func clientActionSetVariables(params interface{}, user **users.User, connID string, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, false, nil
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS
	var ok bool
	var pMap map[string]interface{}
	if pMap, ok = params.(map[string]interface{}); !ok {
		return nil, false, nil
	}
	//SET THE VARIABLES
	userRef.SetVariables(pMap, connID)
	//
	return nil, false, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   FRIENDING ACTIONS   /////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionFriendRequest(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to request a friend")
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, errors.New("Required server features are not enabled")
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}

	requestErr := userRef.FriendRequest(friendName)
	if requestErr != nil {
		return nil, true, requestErr
	}

	//
	return nil, false, nil
}

func clientActionAcceptFriend(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to accept a friend request")
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, errors.New("Required server features are not enabled")
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}

	acceptErr := userRef.AcceptFriendRequest(friendName)
	if acceptErr != nil {
		return nil, true, acceptErr
	}

	//
	return nil, false, nil
}

func clientActionDeclineFriend(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to decline a friend request")
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, errors.New("Required server features are not enabled")
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}

	declineErr := userRef.DeclineFriendRequest(friendName)
	if declineErr != nil {
		return nil, true, declineErr
	}

	//
	return nil, false, nil
}

func clientActionRemoveFriend(params interface{}, user **users.User, clientMux *sync.Mutex) (interface{}, bool, error) {
	(*clientMux).Lock()
	if *user == nil {
		(*clientMux).Unlock()
		return nil, true, errors.New("You must be logged in to remove a friend")
	} else if !(*settings).EnableSqlFeatures {
		(*clientMux).Unlock()
		return nil, true, errors.New("Required server features are not enabled")
	}
	userRef := *user
	(*clientMux).Unlock()
	//GET PARAMS AS A MAP
	var ok bool
	var friendName string

	if friendName, ok = params.(string); !ok {
		return nil, true, errors.New(errorIncorrectFormat)
	}

	removeErr := userRef.RemoveFriend(friendName)
	if removeErr != nil {
		return nil, true, removeErr
	}

	//
	return nil, false, nil
}
