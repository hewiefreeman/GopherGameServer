package database

import (

)

// Represents one of your friends. A friend has a User name, a database index reference, and a status.
// Their status could be FriendStatusRequested, FriendStatusPending, or FriendStatusAccepted (0, 1, or 2). If a User has a Friend
// with the status FriendStatusRequested, they need to accept the request. If a User has a Friend
// with the status FriendStatusPending, that friend has not yet accepted their request. If a User has a Friend
// with the status FriendStatusAccepted, that friend is indeed a friend.
type Friend struct {
	name string
	dbID int
	status int
}

// The three statuses a Friend could be: requested, pending, or accepted (0, 1, and 2). If a User has a Friend
// with the status FriendStatusRequested, they need to accept the request. If a User has a Friend
// with the status FriendStatusPending, that friend has not yet accepted their request. If a User has a Friend
// with the status FriendStatusAccepted, that friend is indeed a friend.
const (
	FriendStatusRequested = iota
	FriendStatusPending
	FriendStatusAccepted
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND FRIEND REQUEST   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to send a
// friend request when using the SQL features.
func FriendRequest(userIndex int, friendIndex int) error {
	_, insertErr := database.Exec("INSERT INTO "+tableFriends+" ("+friendsColumnUser+", "+friendsColumnFriend+", "+friendsColumnStatus+") "+
								"VALUES ("+userIndex+", "+friendIndex+", "+FriendStatusPending+");");
	if(insertErr != nil){ return insertErr; }
	_, insertErr = database.Exec("INSERT INTO "+tableFriends+" ("+friendsColumnUser+", "+friendsColumnFriend+", "+friendsColumnStatus+") "+
								"VALUES ("+friendIndex+", "+userIndex+", "+FriendStatusRequested+");");
	if(insertErr != nil){ return insertErr; }
	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   ACCEPT FRIEND REQUEST   /////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to accept a
// friend request when using the SQL features.
func FriendRequestAccepted(userIndex int, friendIndex int) error {
	_, updateErr := database.Exec("UPDATE "+tableFriends+" SET "+friendsColumnStatus+"="+FriendStatusAccepted+" WHERE ("+friendsColumnUser+"="+userIndex+
							" AND "+friendsColumnFriend+"="+friendIndex+") OR ("+friendsColumnUser+"="+friendIndex+
							" AND "+friendsColumnFriend+"="+userIndex+");");
	if(updateErr != nil){ return updateErr; }
	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   REMOVE FRIEND   /////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to remove a
// friend when using the SQL features.
func RemoveFriend(userIndex int, friendIndex int) error {
	_, updateErr := database.Exec("DELETE FROM "+tableFriends+" WHERE ("+friendsColumnUser+"="+userIndex+" AND "+friendsColumnFriend+"="+friendIndex+") OR ("+
							friendsColumnUser+"="+friendIndex+" AND "+friendsColumnFriend+"="+userIndex+");");
	if(updateErr != nil){ return updateErr; }
	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   GET FRIENDS   /////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the *User.Friends() function
// instead to avoid errors when using the SQL features.
func GetFriends(userIndex int) (map[string]Friend, error) {
	var friends map[string]Friend = make(map[string]Friend);

	//EXECUTE SELECT QUERY
	friendRows, friendRowsErr := database.Query("Select "+friendsColumnFriend+", "+friendsColumnStatus+" FROM "+tableFriends+" WHERE "+friendsColumnUser+"="+userIndex+";");
	if(friendRowsErr != nil){ return friends, friendRowsErr; }
	//
	for friendRows.Next() {
		var friendName string;
		var friendID int;
		var friendStatus int;
		if scanErr := friendRows.Scan(&friendID, &friendStatus); scanErr != nil {
			return friends, scanErr;
		}

		friendInfoRows, friendInfoErr := database.Query("Select "+usersColumnName+" FROM "+tableUsers+" WHERE "+usersColumnID+"="+friendID+" LIMIT 1;");
		if(friendInfoErr != nil){ return friends, friendInfoErr; }
		friendInfoRows.Next();
		if scanErr = friendInfoRows.Scan(&friendName); scanErr != nil {
			return friends, scanErr;
		}
		friendInfoRows.Close();

		friends[friendName] = Friend{name:friendName, dbID: friendID, status: friendStatus};
	}
	friendRows.Close();
	//
	return friends, nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   MAKE A Friend FROM PARAMETERS   /////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func NewFriend(name string, dbID int, status int) Friend {
	return Friend{name: name, dbID: dbID, status: status};
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Friend ATTRIBUTE READERS   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Gets the User name of the Friend.
func (f *Friend) Name() string {
	return f.name;
}

// Gets the database index of the Friend.
func (f *Friend) DatabaseID() int {
	return f.dbID;
}

// Gets the request status of the Friend. Could be either friendStatusRequested or friendStatusAccepted (0 or 1).
func (f *Friend) RequestStatus() int {
	return f.status;
}

// WARNING: This is only meant for internal Gopher Game Server mechanics.
func (f *Friend) SetStatus(status int) {
	f.status = status;
}
