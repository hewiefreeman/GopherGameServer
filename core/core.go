package core

import (
	"github.com/hewiefreeman/GopherGameServer/helpers"
)

var (
	serverStarted bool = false
	serverPaused  bool = false

	serverName        string
	kickOnLogin       bool = false
	sqlFeatures       bool = false
	rememberMe        bool = false
	multiConnect      bool = false
	deleteRoomOnLeave bool = true
)

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
func SettingsSet(kickDups bool, name string, deleteOnLeave bool, sqlFeat bool, remMe bool, multiConn bool) {
	if !serverStarted {
		kickOnLogin = kickDups
		serverName = name
		sqlFeatures = sqlFeat
		rememberMe = remMe
		multiConnect = multiConn
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
//   GET STATE FOR RECOVERY   /////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetState is only for internal Gopher Game Server mechanics.
func GetState() map[string]map[string]interface{} {
	state := make(map[string]map[string]interface{})
	roomsMux.Lock()
	for _, room := range rooms {
		room.mux.Lock()
		state[room.name] = make(map[string]interface{})
		state[room.name]["t"] = room.rType
		state[room.name]["p"] = room.private
		state[room.name]["o"] = room.owner
		state[room.name]["m"] = room.maxUsers
		state[room.name]["i"] = room.inviteList
		state[room.name]["v"] = room.vars
		room.mux.Unlock()
	}
	roomsMux.Unlock()
	//
	return state
}
