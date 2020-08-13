package core

import (
	"errors"
	"github.com/hewiefreeman/GopherGameServer/database"
	"github.com/hewiefreeman/GopherGameServer/helpers"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND A FRIEND REQUEST   /////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// FriendRequest sends a friend request to another User by their name.
func (u *User) FriendRequest(friendName string) error {
	if _, ok := u.friends[friendName]; ok {
		return errors.New("The user '" + friendName + "' cannot be requested as a friend")
	}
	//CHECK IF FRIEND IS ONLINE & GET DATABASE ID
	friend, friendErr := GetUser(friendName)
	var friendOnline bool = false
	var friendID int
	if friendErr != nil {
		//GET FRIEND'S DATABASE ID FROM database PACKAGE
		friendID, friendErr = database.GetUserDatabaseIndex(friendName)
		if friendErr != nil {
			return errors.New("The user '" + friendName + "' does not exist")
		}
	} else {
		friendID = friend.databaseID
		friendOnline = true
	}

	//ADD REQUESTED FRIEND FOR USER
	u.mux.Lock()
	u.friends[friendName] = database.NewFriend(friendName, friendID, database.FriendStatusPending)
	u.mux.Unlock()

	//ADD REQUESTED FRIEND FOR FRIEND
	if friendOnline {
		friend.mux.Lock()
		friend.friends[u.name] = database.NewFriend(u.name, u.databaseID, database.FriendStatusRequested)
		friend.mux.Unlock()
	}

	//MAKE THE FRIEND REQUEST ON DATABASE
	friendingErr := database.FriendRequest(u.databaseID, friendID)
	if friendingErr != nil {
		return errors.New("Unexpected friend error")
	}

	//SEND A FRIEND REQUEST TO THE USER IF THEY ARE ONLINE
	if friendOnline {
		message := map[string]map[string]interface{}{
			helpers.ServerActionFriendRequest: {
				"n": u.name,
			},
		}
		friend.mux.Lock()
		for _, conn := range friend.conns {
			(*conn).socket.WriteJSON(message)
		}
		friend.mux.Unlock()
	}

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionFriendRequest, friendName, helpers.NoError())
	u.mux.Lock()
	for _, conn := range u.conns {
		(*conn).socket.WriteJSON(clientResp)
	}
	u.mux.Unlock()

	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ACCEPT A FRIEND REQUEST   ///////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// AcceptFriendRequest accepts a friend request from another User by their name.
func (u *User) AcceptFriendRequest(friendName string) error {
	if _, ok := u.friends[friendName]; !ok {
		return errors.New("The user '" + friendName + "' has not requested you as a friend")
	} else if (u.friends[friendName]).RequestStatus() != database.FriendStatusRequested {
		return errors.New("The user '" + friendName + "' cannot be accepted as a friend")
	}
	//CHECK IF FRIEND IS ONLINE & GET DATABASE ID
	friend, friendErr := GetUser(friendName)
	var friendOnline bool = false
	var friendID int
	if friendErr != nil {
		//GET FRIEND'S DATABASE ID FROM database PACKAGE
		friendID, friendErr = database.GetUserDatabaseIndex(friendName)
		if friendErr != nil {
			return errors.New("The user '" + friendName + "' does not exist")
		}
	} else {
		friendID = friend.databaseID
		friendOnline = true
	}

	//ACCEPT FRIEND FOR USER
	u.mux.Lock()
	u.friends[friendName].SetStatus(database.FriendStatusAccepted)
	u.mux.Unlock()
	//ACCEPT FRIEND FOR FRIEND
	if friendOnline {
		friend.mux.Lock()
		friend.friends[u.name].SetStatus(database.FriendStatusAccepted)
		friend.mux.Unlock()
	}
	//UPDATE FRIENDS ON DATABASE
	friendingErr := database.FriendRequestAccepted(u.databaseID, friendID)
	if friendingErr != nil {
		return errors.New("Unexpected friend error")
	}

	//SEND ACCEPT MESSAGE TO THE USER IF THEY ARE ONLINE
	var fStatus int = StatusOffline
	if friendOnline {
		message := map[string]map[string]interface{}{
			helpers.ServerActionFriendAccept: {
				"n": u.name,
				"s": u.status,
			},
		}
		friend.mux.Lock()
		for _, conn := range friend.conns {
			(*conn).socket.WriteJSON(message)
		}
		fStatus = friend.status
		friend.mux.Unlock()
	}

	//MAKE RESPONSE
	responseMap := map[string]interface{}{
		"n": friendName,
		"s": fStatus,
	}

	//SEND RESPONSE TO ALL CLIENT CONNECTIONS
	clientResp := helpers.MakeClientResponse(helpers.ClientActionAcceptFriend, responseMap, helpers.NoError())
	u.mux.Lock()
	for _, conn := range u.conns {
		(*conn).socket.WriteJSON(clientResp)
	}
	u.mux.Unlock()

	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Decline friend request   ////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// DeclineFriendRequest declines a friend request from another User by their name.
func (u *User) DeclineFriendRequest(friendName string) error {
	if _, ok := u.friends[friendName]; !ok {
		return errors.New("The user '" + friendName + "' has not requested you as a friend")
	} else if u.friends[friendName].RequestStatus() != database.FriendStatusRequested {
		return errors.New("The user '" + friendName + "' cannot be declined as a friend")
	}
	//CHECK IF FRIEND IS ONLINE & GET DATABASE ID
	friend, friendErr := GetUser(friendName)
	var friendOnline bool = false
	var friendID int
	if friendErr != nil {
		//GET FRIEND'S DATABASE ID FROM database PACKAGE
		friendID, friendErr = database.GetUserDatabaseIndex(friendName)
		if friendErr != nil {
			return errors.New("The user '" + friendName + "' does not exist")
		}
	} else {
		friendID = friend.databaseID
		friendOnline = true
	}

	//DELETE THE Users' Friends
	u.mux.Lock()
	delete(u.friends, friendName)
	u.mux.Unlock()
	//ACCEPT FRIEND FOR FRIEND
	if friendOnline {
		friend.mux.Lock()
		delete(friend.friends, u.name)
		friend.mux.Unlock()
	}

	//UPDATE FRIENDS ON DATABASE
	removeErr := database.RemoveFriend(u.databaseID, friendID)
	if removeErr != nil {
		return errors.New("Unexpected friend error")
	}

	//SEND A FRIEND REQUEST TO THE USER IF THEY ARE ONLINE
	if friendOnline {
		message := map[string]map[string]interface{}{
			helpers.ServerActionFriendRemove: {
				"n": u.name,
			},
		}
		friend.mux.Lock()
		for _, conn := range friend.conns {
			(*conn).socket.WriteJSON(message)
		}
		friend.mux.Unlock()
	}

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionDeclineFriend, friendName, helpers.NoError())
	u.mux.Lock()
	for _, conn := range u.conns {
		(*conn).socket.WriteJSON(clientResp)
	}
	u.mux.Unlock()

	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Remove a friend   ///////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// RemoveFriend removes a friend from this this User and this User from the friend's Friend list.
func (u *User) RemoveFriend(friendName string) error {
	if _, ok := u.friends[friendName]; !ok {
		return errors.New("The user '" + friendName + "' is not your friend")
	} else if u.friends[friendName].RequestStatus() != database.FriendStatusAccepted {
		return errors.New("The user '" + friendName + "' cannot be removed as a friend")
	}
	//CHECK IF FRIEND IS ONLINE & GET DATABASE ID
	friend, friendErr := GetUser(friendName)
	var friendOnline bool = false
	var friendID int
	if friendErr != nil {
		//GET FRIEND'S DATABASE ID FROM database PACKAGE
		friendID, friendErr = database.GetUserDatabaseIndex(friendName)
		if friendErr != nil {
			return errors.New("The user '" + friendName + "' does not exist")
		}
	} else {
		friendID = friend.databaseID
		friendOnline = true
	}

	//DELETE THE Users' Friends
	u.mux.Lock()
	delete(u.friends, friendName)
	u.mux.Unlock()
	//ACCEPT FRIEND FOR FRIEND
	if friendOnline {
		friend.mux.Lock()
		delete(friend.friends, u.name)
		friend.mux.Unlock()
	}

	//UPDATE FRIENDS ON DATABASE
	removeErr := database.RemoveFriend(u.databaseID, friendID)
	if removeErr != nil {
		return errors.New("Unexpected friend error")
	}

	//SEND A FRIEND REQUEST TO THE USER IF THEY ARE ONLINE
	if friendOnline {
		message := map[string]map[string]interface{}{
			helpers.ServerActionFriendRemove: {
				"n": u.name,
			},
		}
		friend.mux.Lock()
		for _, conn := range friend.conns {
			(*conn).socket.WriteJSON(message)
		}
		friend.mux.Unlock()
	}

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionRemoveFriend, friendName, helpers.NoError())
	u.mux.Lock()
	for _, conn := range u.conns {
		(*conn).socket.WriteJSON(clientResp)
	}
	u.mux.Unlock()

	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Send message to all friends   ///////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func (u *User) sendToFriends(message interface{}) {
	for key, val := range u.friends {
		if val.RequestStatus() == database.FriendStatusAccepted {
			friend, friendErr := GetUser(key)
			if friendErr == nil {
				friend.mux.Lock()
				for _, friendConn := range friend.conns {
					(*friendConn).socket.WriteJSON(message)
				}
				friend.mux.Unlock()
			}
		}
	}
}