package sql

import (
	//"database/sql"
)

// CONFIGURES THE DATABASE FOR Gopher Game Server USAGE
func setUp() error {
	//CHECK IF THE TABLE users HAS BEEN MADE
	_, checkErr := database.Exec("SELECT "+usersColumnName+" FROM "+tableUsers+" WHERE "+usersColumnID+"=1;");
	if(checkErr != nil){
		//MAKE THE users TABLE
		_, createErr := database.Exec("CREATE TABLE "+tableUsers+" ("+
								usersColumnID+" INTEGER NOT NULL AUTO_INCREMENT, "+
								usersColumnName+" varchar(255) NOT NULL, "+
								usersColumnPassword+" varchar(255) NOT NULL, "+
								"PRIMARY KEY ("+usersColumnID+")"+
								");");
		if(createErr != nil){ return createErr; }

		//ADJUST AUTO_INCREMENT TO 1
		_, adjustErr := database.Exec("ALTER TABLE "+tableUsers+" AUTO_INCREMENT=1;");
		if(adjustErr != nil){ return adjustErr; }

		//INSERT startUpTestUser FOR FUTURE TESTS
		_, insertErr := database.Exec("INSERT INTO "+tableUsers+" ("+usersColumnName+", "+usersColumnPassword+") "+
								"VALUES (\"startUpTestUser\", \"startUpTestUser\");");
		if(insertErr != nil){ return insertErr; }

		//MAKE THE friends TABLE
		_, friendsErr := database.Exec("CREATE TABLE "+tableFriends+" ("+
								friendsColumnUser+" INTEGER NOT NULL, "+
								friendsColumnFriend+" INTEGER NOT NULL, "+
								friendsColumnStatus+" INTEGER NOT NULL, "+
								"FOREIGN KEY("+friendsColumnUser+") REFERENCES "+tableUsers+"("+usersColumnID+"), "+
								"FOREIGN KEY("+friendsColumnFriend+") REFERENCES "+tableUsers+"("+usersColumnID+")"+
								");");
		if(friendsErr != nil){ return friendsErr; }


	}

	//
	return nil;
}
