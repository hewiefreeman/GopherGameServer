package helpers

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
	ServerActionWebRTCOffer                = "wo"
)

// MakeClientResponse is used for Gopher Game Server inner mechanics only.
func MakeClientResponse(action string, responseVal interface{}, err GopherError) map[string]map[string]interface{} {
	var response map[string]map[string]interface{}
	if err.ID != 0 {
		response = map[string]map[string]interface{}{
			ServerActionClientActionResponse: {
				"a": action,
				"e": map[string]interface{}{
					"m":  err.Message,
					"id": err.ID,
				},
			},
		}
	} else {
		response = map[string]map[string]interface{}{
			ServerActionClientActionResponse: {
				"a": action,
				"r": responseVal,
			},
		}
	}

	//
	return response
}
