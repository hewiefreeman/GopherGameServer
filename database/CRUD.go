package database

import (
	"errors"
)

// Insert data into your custom table. The data map must have keys that match your custom table column names, and the
// attached interface will be checked for the correct data type and inserted accordingly.
func Insert(tableName string, data map[string]interface{}) error {
	if(tableName == tableUsers || tableName == tableFriends){
		return errors.New("The "+tableUsers+" and "+tableFriends+" tables are for Gopher Game Server mechanics only.");
	}else if(len(tableName) == 0){
		return errors.New("database.Insert() requires a table name");
	}else if(len(data) == 0){
		return errors.New("database.Insert() requires data");
	}

	//CHECK DATA COLUMN NAMES AND TYPES

	//
	return nil;
}
