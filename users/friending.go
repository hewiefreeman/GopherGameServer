package users

import (
	"github.com/hewiefreeman/GopherGameServer/database"
	"errors"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND A FRIEND REQUEST   /////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func (u *User) FriendRequest(friendName string) error {
	if _, ok := u.friends[friendName]; ok {
		return errors.New("The user '"+friendName+"' cannot be requested as a friend");
	}
	//CHECK IF FRIEND IS ONLINE & GET DATABASE ID
	friend, friendErr := Get(friendName);
	var friendOnline bool = false;
	var friendID int;
	if(friendErr != nil){
		//GET FRIEND'S DATABASE ID FROM database PACKAGE
		friendID, friendErr = database.GetUserDatabaseIndex(friendName);
		if(friendErr != nil){ return errors.New("The user '"+friendName+"' does not exist"); }
	}else{
		friendID = friend.databaseID;
		friendOnline = true;
	}

	//ADD TO THE Users' Friends
	response := usersActionChan.Execute(addFriend, []interface{}{u.name, u.databaseID, friendName, friendID});
	if(len(response) == 0 || response[0] != nil){
		return response[0].(error);
	}

	//MAKE THE FRIEND REQUEST ON DATABASE
	friendingErr := database.FriendRequest(u.databaseID, friendID);
	if(friendingErr != nil){ return errors.New("Unexpected friend error"); }

	//SEND A FRIEND REQUEST TO THE USER IF THEY ARE ONLINE
	if(friendOnline){
		message := make(map[string]interface{});
		message["f"] = make(map[string]interface{});
		message["f"].(map[string]interface{})["n"] = u.name;
		friend.socket.WriteJSON(message);
	}

	//
	return nil;
}

func addFriend(params []interface{}) []interface{} {
	userName, userID, friendName, friendID := params[0].(string), params[1].(int), params[2].(string), params[3].(int);
	//ADD PENDING FRIEND FOR USER
	if _, ok := users[userName]; ok {
		(*users[userName]).friends[friendName] = database.NewFriend(friendName, friendID, database.FriendStatusPending);
	}else{
		return []interface{}{ errors.New("User '"+userName+"' is not logged in") };
	}
	//ADD REQUESTED FRIEND FOR FRIEND
	if _, ok = users[friendName]; ok {
		(*users[friendName]).friends[userName] = database.NewFriend(userName, userID, database.FriendStatusRequested);
	}
	//
	return []interface{}{nil};
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ACCEPT A FRIEND REQUEST   ///////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func (u *User) AcceptFriendRequest(friendName string) error {
	if _, ok := u.friends[friendName]; !ok {
		return errors.New("The user '"+friendName+"' has not requested you as a friend");
	}else if(u.friends[friendName].RequestStatus() != database.FriendStatusRequested){
		return errors.New("The user '"+friendName+"' cannot be accepted as a friend");
	}
	//CHECK IF FRIEND IS ONLINE & GET DATABASE ID
	friend, friendErr := Get(friendName);
	var friendOnline bool = false;
	var friendID int;
	if(friendErr != nil){
		//GET FRIEND'S DATABASE ID FROM database PACKAGE
		friendID, friendErr = database.GetUserDatabaseIndex(friendName);
		if(friendErr != nil){ return errors.New("The user '"+friendName+"' does not exist"); }
	}else{
		friendID = friend.databaseID;
		friendOnline = true;
	}

	//UPDATE THE Users' Friends
	response := usersActionChan.Execute(friendAccepted, []interface{}{u.name, friendName});
	if(len(response) == 0 || response[0] != nil){
		return response[0].(error);
	}

	//UPDATE FRIENDS ON DATABASE
	friendingErr := database.FriendRequestAccepted(u.databaseID, friendID);
	if(friendingErr != nil){ return errors.New("Unexpected friend error"); }

	//SEND A FRIEND REQUEST TO THE USER IF THEY ARE ONLINE
	if(friendOnline){
		message := make(map[string]interface{});
		message["fa"] = make(map[string]interface{});
		message["fa"].(map[string]interface{})["n"] = u.name;
		friend.socket.WriteJSON(message);
	}

	//
	return nil;
}

func friendAccepted(params []interface{}) []interface{} {
	userName, friendName := params[0].(string), params[1].(string);
	//ACCEPT FRIEND FOR USER
	if user, ok := users[userName]; ok {
		if _, ok = user.friends[friendName]; !ok {
			return []interface{}{ errors.New("Unexpected friend error") };
		}
		//ACCEPT FRIEND FOR FRIEND
		if user, ok = users[friendName]; ok {
			if _, ok = user.friends[userName]; !ok {
				return []interface{}{ errors.New("Unexpected friend error") };
			}
			(*users[friendName]).friends[userName].SetStatus(database.FriendStatusAccepted);
		}
		(*users[userName]).friends[friendName].SetStatus(database.FriendStatusAccepted);
	}else{
		return []interface{}{ errors.New("User '"+userName+"' is not logged in") };
	}
	//
	return []interface{}{nil};
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   DECLINE A FRIEND REQUEST   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func (u *User) DeclineFriendRequest(friendName string) error {
	if _, ok := u.friends[friendName]; !ok {
		return errors.New("The user '"+friendName+"' has not requested you as a friend");
	}else if(u.friends[friendName].RequestStatus() != database.FriendStatusRequested){
		return errors.New("The user '"+friendName+"' cannot be declined as a friend");
	}
	//CHECK IF FRIEND IS ONLINE & GET DATABASE ID
	friend, friendErr := Get(friendName);
	var friendOnline bool = false;
	var friendID int;
	if(friendErr != nil){
		//GET FRIEND'S DATABASE ID FROM database PACKAGE
		friendID, friendErr = database.GetUserDatabaseIndex(friendName);
		if(friendErr != nil){ return errors.New("The user '"+friendName+"' does not exist"); }
	}else{
		friendID = friend.databaseID;
		friendOnline = true;
	}

	//DELETE THE Users' Friends
	response := usersActionChan.Execute(removeFriends, []interface{}{u.name, friendName});
	if(len(response) == 0 || response[0] != nil){
		return response[0].(error);
	}

	//UPDATE FRIENDS ON DATABASE
	removeErr := database.RemoveFriend(u.databaseID, friendID);
	if(removeErr != nil){ return errors.New("Unexpected friend error"); }

	//SEND A FRIEND REQUEST TO THE USER IF THEY ARE ONLINE
	if(friendOnline){
		message := make(map[string]interface{});
		message["fr"] = make(map[string]interface{});
		message["fr"].(map[string]interface{})["n"] = u.name;
		friend.socket.WriteJSON(message);
	}

	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   REMOVE A FRIEND   ///////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func (u *User) RemoveFriend(friendName string) error {
	if _, ok := u.friends[friendName]; !ok {
		return errors.New("The user '"+friendName+"' is not your friend");
	}else if(u.friends[friendName].RequestStatus() != database.FriendStatusAccepted){
		return errors.New("The user '"+friendName+"' cannot be removed as a friend");
	}
	//CHECK IF FRIEND IS ONLINE & GET DATABASE ID
	friend, friendErr := Get(friendName);
	var friendOnline bool = false;
	var friendID int;
	if(friendErr != nil){
		//GET FRIEND'S DATABASE ID FROM database PACKAGE
		friendID, friendErr = database.GetUserDatabaseIndex(friendName);
		if(friendErr != nil){ return errors.New("The user '"+friendName+"' does not exist"); }
	}else{
		friendID = friend.databaseID;
		friendOnline = true;
	}

	//DELETE THE Users' Friends
	response := usersActionChan.Execute(removeFriends, []interface{}{u.name, friendName});
	if(len(response) == 0 || response[0] != nil){
		return response[0].(error);
	}

	//UPDATE FRIENDS ON DATABASE
	removeErr := database.RemoveFriend(u.databaseID, friendID);
	if(removeErr != nil){ return errors.New("Unexpected friend error"); }

	//SEND A FRIEND REQUEST TO THE USER IF THEY ARE ONLINE
	if(friendOnline){
		message := make(map[string]interface{});
		message["fr"] = make(map[string]interface{});
		message["fr"].(map[string]interface{})["n"] = u.name;
		friend.socket.WriteJSON(message);
	}

	//
	return nil;
}

func removeFriends(params []interface{}) []interface{} {
	userName, friendName := params[0].(string), params[1].(string);
	//ACCEPT FRIEND FOR USER
	if user, ok := users[userName]; ok {
		if _, ok = user.friends[friendName]; !ok {
			return []interface{}{ errors.New("Unexpected friend error") };
		}
		//ACCEPT FRIEND FOR FRIEND
		if user, ok = users[friendName]; ok {
			if _, ok = user.friends[userName]; !ok {
				return []interface{}{ errors.New("Unexpected friend error") };
			}
			delete((*users[friendName]).friends, userName);
		}
		delete((*users[userName]).friends, friendName);
	}else{
		return []interface{}{ errors.New("User '"+userName+"' is not logged in") };
	}
	//
	return []interface{}{nil};
}
