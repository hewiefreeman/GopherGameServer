// Package core contains all the tools to make and work with Users and Rooms.
//
// A User is a client who has successfully logged into the server. You can think of clients who are not attached to a User
// as, for instance, someone in the login screen, but are still connected to the server. A client doesn't
// have to be a User to be able to call your CustomClientActions, so keep that in mind when making them (Refer to the Usage for CustomClientActions).
//
// Users have their own variables which can be accessed and changed anytime. A User variable can
// be anything compatible with interface{}, so pretty much anything.
//
// A Room represents a place on the server where a User can join other Users. Rooms can either be public or private. Private Rooms must be assigned an "owner", which is the name of a User, or the ServerName
// from ServerSettings. The server's name that will be used for ownership of private Rooms can be set with the ServerSettings
// option ServerName when starting the server. Though keep in mind, setting the ServerName in ServerSettings will prevent a User who wants to go by that name
// from logging in. Public Rooms will accept a join request from any User, and private Rooms will only
// accept a join request from someone who is on it's invite list. Only the owner of the Room or the server itself can invite
// Users to a private Room. But remember, just because a User owns a private room doesn't mean the server cannot also invite
// to the room via *Room.AddInvite() function.
//
// Rooms have their own variables which can be accessed and changed anytime. Like User variables, a Room variable can
// be anything compatible with interface{}.
package core

import (
	"github.com/hewiefreeman/GopherGameServer/helpers"
)

var (
	serverStarted bool
	serverPaused  bool

	serverName        string
	kickOnLogin       bool
	sqlFeatures       bool
	rememberMe        bool
	multiConnect      bool
	maxUserConns      uint8
	deleteRoomOnLeave bool = true
)

// RoomRecoveryState is used internally for persisting room states on shutdown.
type RoomRecoveryState struct {
	T string                 // rType
	P bool                   // private
	O string                 // owner
	M int                    // maxUsers
	I []string               // inviteList
	V map[string]interface{} // vars
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER STARTUP FUNCTIONS   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// SetServerStarted is for Gopher Game Server internal mechanics only.
func SetServerStarted(val bool) {
	if !serverStarted {
		serverStarted = val
	}
}

// SettingsSet is for Gopher Game Server internal mechanics only.
func SettingsSet(kickDups bool, name string, deleteOnLeave bool, sqlFeat bool, remMe bool, multiConn bool, maxConns uint8) {
	if !serverStarted {
		kickOnLogin = kickDups
		serverName = name
		sqlFeatures = sqlFeat
		rememberMe = remMe
		multiConnect = multiConn
		maxUserConns = maxConns
		deleteRoomOnLeave = deleteOnLeave
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER PAUSE AND RESUME   ///////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Pause is only for internal Gopher Game Server mechanics.
func Pause() {
	if !serverPaused {
		serverPaused = true

		//
		clientResp := helpers.MakeClientResponse(helpers.ClientActionLogout, nil, helpers.NoError())
		usersMux.Lock()
		for _, user := range users {
			user.mux.Lock()
			for connID, conn := range user.conns {
				//REMOVE CONNECTION FROM THEIR ROOM
				currRoom := conn.room
				if currRoom != nil && currRoom.Name() != "" {
					user.mux.Unlock()
					currRoom.RemoveUser(user, connID)
					user.mux.Lock()
				}

				//LOG CONNECTION OUT
				conn.clientMux.Lock()
				if *(conn.user) != nil {
					*(conn.user) = nil
				}
				conn.clientMux.Unlock()

				//SEND LOG OUT MESSAGE
				conn.socket.WriteJSON(clientResp)
			}
			user.mux.Unlock()
		}
		users = make(map[string]*User)
		usersMux.Unlock()
	}
}

// Resume is only for internal Gopher Game Server mechanics.
func Resume() {
	if serverPaused {
		serverPaused = false
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   GET STATES FOR GENERATING RECOVERY FILE   ////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetRoomsState is only for internal Gopher Game Server mechanics.
func GetRoomsState() map[string]RoomRecoveryState {
	state := make(map[string]RoomRecoveryState)
	roomsMux.Lock()
	for _, room := range rooms {
		room.mux.Lock()
		state[room.name] = RoomRecoveryState{
			T: room.rType,
			P: room.private,
			O: room.owner,
			M: room.maxUsers,
			I: room.inviteList,
			V: room.vars,
		}
		room.mux.Unlock()
	}
	roomsMux.Unlock()
	//
	return state
}
