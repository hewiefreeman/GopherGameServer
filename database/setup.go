package database

import (
	"fmt"
	"strconv"
)

// CONFIGURES THE DATABASE FOR Gopher Game Server USAGE
func setUp() error {
	//CHECK IF THE TABLE users HAS BEEN MADE
	_, checkErr := database.Exec("SELECT " + usersColumnName + " FROM " + tableUsers + " WHERE " + usersColumnID + "=1;")
	if checkErr != nil {
		fmt.Println("Making all tables...")
		//MAKE THE users TABLE QUERY
		createQuery := "CREATE TABLE " + tableUsers + " (" +
			usersColumnID + " INTEGER NOT NULL AUTO_INCREMENT, " +
			usersColumnName + " VARCHAR(255) UNIQUE NOT NULL, " +
			usersColumnPassword + " VARCHAR(255) NOT NULL, "

		//APPEND custom AccountInfoColumn ITEMS
		for key, val := range customAccountInfo {
			createQuery = createQuery + key + " " + dataTypes[val.dataType]
			//CHECK IF NEEDS maxSize/precision
			if isSizeDataType(val.dataType) {
				createQuery = createQuery + "(" + strconv.Itoa(val.maxSize) + ")"
			} else if isPrecisionDataType(val.dataType) {
				createQuery = createQuery + "(" + strconv.Itoa(val.maxSize) + ", " + strconv.Itoa(val.precision) + ")"
			}
			//CHECK IF UNIQUE
			if val.unique {
				createQuery = createQuery + " UNIQUE"
			}
			//CHECK IF NOT NULL
			if val.notNull {
				createQuery = createQuery + " NOT NULL, "
			} else {
				createQuery = createQuery + ", "
			}
		}

		createQuery = createQuery + "PRIMARY KEY (" + usersColumnID + "));"

		//EXECUTE users TABLE QUERY
		_, createErr := database.Exec(createQuery)
		if createErr != nil {
			return createErr
		}

		//ADJUST AUTO_INCREMENT TO 1
		_, adjustErr := database.Exec("ALTER TABLE " + tableUsers + " AUTO_INCREMENT=1;")
		if adjustErr != nil {
			return adjustErr
		}

		//MAKE THE friends TABLE
		_, friendsErr := database.Exec("CREATE TABLE " + tableFriends + " (" +
			friendsColumnUser + " INTEGER NOT NULL, " +
			friendsColumnFriend + " INTEGER NOT NULL, " +
			friendsColumnStatus + " INTEGER NOT NULL, " +
			"FOREIGN KEY(" + friendsColumnUser + ") REFERENCES " + tableUsers + "(" + usersColumnID + "), " +
			"FOREIGN KEY(" + friendsColumnFriend + ") REFERENCES " + tableUsers + "(" + usersColumnID + ")" +
			");")
		if friendsErr != nil {
			return friendsErr
		}

		if rememberMe {
			//MAKE THE autologs TABLE
			_, friendsErr = database.Exec("CREATE TABLE " + tableAutologs + " (" +
				autologsColumnID + " INTEGER NOT NULL, " +
				autologsColumnDevicePass + " VARCHAR(255) NOT NULL, " +
				autologsColumnDeviceTag + " VARCHAR(255) NOT NULL, " +
				");")
			if friendsErr != nil {
				return friendsErr
			}
		}

	} else {
		//CHECK IF THERE ARE ANY NEW custom AccountInfoColumn ITEMS
		query := "ALTER TABLE " + tableUsers + " "
		var execQuery bool = false
		//
		for key, val := range customAccountInfo {
			//CHECK IF COLUMN EXISTS
			checkRows, err := database.Query("SHOW COLUMNS FROM " + tableUsers + " LIKE '" + key + "';")
			if err != nil {
				return err
			}
			//
			checkRows.Next()
			_, colsErr := checkRows.Columns()
			if colsErr != nil {
				fmt.Println("Adding AccountInfoColumn '" + key + "'...")
				//THIS customAccountInfo COLUMN DOES NOT EXIST... YET, MY NERD.
				query = query + "ADD COLUMN " + key + " " + dataTypes[val.dataType]
				if isSizeDataType(val.dataType) {
					query = query + "(" + strconv.Itoa(val.maxSize) + ")"
				} else if isPrecisionDataType(val.dataType) {
					query = query + "(" + strconv.Itoa(val.maxSize) + ", " + strconv.Itoa(val.precision) + ")"
				}
				//CHECK IF UNIQUE
				if val.unique {
					query = query + " UNIQUE"
				}
				//CHECK IF NOT NULL
				if val.notNull {
					query = query + " NOT NULL, "
				} else {
					query = query + ", "
				}
				execQuery = true
			}
			checkRows.Close()
		}
		if execQuery {
			//MAKE THE NEW COLUMNS
			query = query[0:len(query)-2] + ";"
			_, colsErr := database.Exec(query)
			if colsErr != nil {
				return colsErr
			}
		}

		if rememberMe {
			//CHECK IF THE autologs TABLE HAS BEEN MADE
			_, checkErr := database.Exec("SELECT " + autologsColumnID + " FROM " + tableAutologs + " WHERE " + autologsColumnID + "=1;")
			if checkErr != nil {
				fmt.Println("Making autologs table...")
				//MAKE THE autologs TABLE
				_, friendsErr := database.Exec("CREATE TABLE " + tableAutologs + " (" +
					autologsColumnID + " INTEGER NOT NULL, " +
					autologsColumnDevicePass + " VARCHAR(255) NOT NULL, " +
					autologsColumnDeviceTag + " VARCHAR(255) NOT NULL" +
					");")
				if friendsErr != nil {
					return friendsErr
				}
			}
		}
	}
	//MAKE SURE THE customLoginColumn IS UNIQUE IF SET
	if len(customLoginColumn) > 0 {
		_, alterErr := database.Exec("ALTER TABLE " + tableUsers + " ADD UNIQUE (" + customLoginColumn + ");")
		if alterErr != nil {
			return alterErr
		}
	}

	//
	return nil
}
