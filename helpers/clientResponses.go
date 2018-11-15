package helpers

import ()

//BUILT-IN CLIENT ACTION/RESPONSE MESSAGE TYPES
const (
	ClientActionSignup            = "s"
	ClientActionDeleteAccount     = "d"
	ClientActionChangePassword    = "pc"
	ClientActionChangeAccountInfo = "ic"
	ClientActionLogin             = "li"
	ClientActionLogout            = "lo"
	ClientActionJoinRoom          = "j"
	ClientActionLeaveRoom         = "lr"
	ClientActionCreateRoom        = "r"
	ClientActionDeleteRoom        = "rd"
	ClientActionRoomInvite        = "i"
	ClientActionRevokeInvite      = "ri"
	ClientActionChatMessage       = "c"
	ClientActionPrivateMessage    = "p"
	ClientActionVoiceStream       = "v"
	ClientActionChangeStatus      = "sc"
	ClientActionCustomAction      = "a"
	ClientActionFriendRequest     = "f"
	ClientActionAcceptFriend      = "fa"
	ClientActionDeclineFriend     = "fd"
	ClientActionRemoveFriend      = "fr"
	ClientActionSetVariable       = "vs"
	ClientActionSetVariables      = "vx"
)

//BUILT-IN SERVER ACTION RESPONSES
const (
	ServerActionClientActionResponse       = "c"
	ServerActionCustomClientActionResponse = "a"
	ServerActionDataMessage                = "d"
	ServerActionPrivateMessage             = "p"
	ServerActionRoomMessage                = "m"
	ServerActionUserEnter                  = "e"
	ServerActionUserLeave                  = "x"
	ServerActionVoiceStream                = "v"
	ServerActionVoicePing                  = "vp"
	ServerActionRoomInvite                 = "i"
	ServerActionFriendRequest              = "f"
	ServerActionFriendAccept               = "fa"
	ServerActionFriendRemove               = "fr"
	ServerActionFriendStatusChange         = "fs"
	ServerActionRequestDeviceTag           = "t"
	ServerActionSetDeviceTag               = "ts"
	ServerActionSetAutoLoginPass           = "ap"
	ServerActionAutoLoginFailed            = "af"
	ServerActionAutoLoginNotFiled          = "ai"
)

// MakeClientResponse is used for Gopher Game Server inner mechanics only.
func MakeClientResponse(action string, responseVal interface{}, err error) map[string]interface{} {
	var response map[string]interface{}
	if err != nil {
		response = make(map[string]interface{})
		response[ServerActionClientActionResponse] = make(map[string]interface{})
		response[ServerActionClientActionResponse].(map[string]interface{})["a"] = action
		response[ServerActionClientActionResponse].(map[string]interface{})["e"] = err.Error()
	} else {
		response = make(map[string]interface{})
		response[ServerActionClientActionResponse] = make(map[string]interface{})
		response[ServerActionClientActionResponse].(map[string]interface{})["a"] = action
		response[ServerActionClientActionResponse].(map[string]interface{})["r"] = responseVal
	}

	//
	return response
}
