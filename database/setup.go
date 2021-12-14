package database

import (
	"fmt"
	"strconv"
)

// Configures the SQL database for Gopher Game Server
func setUp() error {
	// Check if the users table has been created
	_, checkErr := database.Exec("SELECT " + usersColumnName + " FROM " + tableUsers + " WHERE " + usersColumnID + "=1;")
	if checkErr != nil {
		fmt.Println("Creating \"" + tableUsers + "\" table...")
		// Make the users table
		if cErr := createUserTableSQL(); cErr != nil {
			return cErr
		}
	}
	// Check for new custom AccountInfoColumn items
	if newItemsErr := addNewCustomItemsSQL(); newItemsErr != nil {
		return newItemsErr
	}

	if rememberMe {
		// Check if autologs table has been made
		_, checkErr := database.Exec("SELECT " + autologsColumnID + " FROM " + tableAutologs + " WHERE " + autologsColumnID + "=1;")
		if checkErr != nil {
			fmt.Println("Making autologs table...")
			if cErr := createAutologsTableSQL(); cErr != nil {
				return cErr
			}
		}
	}
	// Make sure customLoginColumn is unique if it is set
	if len(customLoginColumn) > 0 {
		_, alterErr := database.Exec("ALTER TABLE " + tableUsers + " ADD UNIQUE (" + customLoginColumn + ");")
		if alterErr != nil {
			return alterErr
		}
	}

	//
	return nil
}

func createUserTableSQL() error {
	createQuery := "CREATE TABLE " + tableUsers + " (" +
	usersColumnID + " INTEGER NOT NULL AUTO_INCREMENT, " +
	usersColumnName + " VARCHAR(255) UNIQUE NOT NULL, " +
	usersColumnPassword + " VARCHAR(255) NOT NULL, "

	// Append custom AccountInfoColumn items
	for key, val := range customAccountInfo {
		createQuery = createQuery + key + " " + dataTypes[val.dataType]
		// Check for maxSize/precision
		if isSizeDataType(val.dataType) {
			createQuery = createQuery + "(" + strconv.Itoa(val.maxSize) + ")"
		} else if isPrecisionDataType(val.dataType) {
			createQuery = createQuery + "(" + strconv.Itoa(val.maxSize) + ", " + strconv.Itoa(val.precision) + ")"
		}
		// Check for unique
		if val.unique {
			createQuery = createQuery + " UNIQUE"
		}
		// Check for not-null
		if val.notNull {
			createQuery = createQuery + " NOT NULL, "
		} else {
			createQuery = createQuery + ", "
		}
	}

	createQuery = createQuery + "PRIMARY KEY (" + usersColumnID + "));"

	// Execute users table query
	_, createErr := database.Exec(createQuery)
	if createErr != nil {
		return createErr
	}

	// Adjust auto-increment to 1
	_, adjustErr := database.Exec("ALTER TABLE " + tableUsers + " AUTO_INCREMENT=1;")
	if adjustErr != nil {
		return adjustErr
	}

	// Make friends table
	if _, friendsErr := database.Exec("CREATE TABLE " + tableFriends + " (" +
		friendsColumnUser + " INTEGER NOT NULL, " +
		friendsColumnFriend + " INTEGER NOT NULL, " +
		friendsColumnStatus + " INTEGER NOT NULL" +
		");"); friendsErr != nil {

		return friendsErr
	}

	if rememberMe {
		if cErr := createAutologsTableSQL(); cErr != nil {
			return cErr
		}
	}

	return nil
}

func createAutologsTableSQL() error {
	if _, aErr := database.Exec("CREATE TABLE " + tableAutologs + " (" +
		autologsColumnID + " INTEGER NOT NULL, " +
		autologsColumnDevicePass + " VARCHAR(255) NOT NULL, " +
		autologsColumnDeviceTag + " VARCHAR(255) NOT NULL, " +
		");"); aErr != nil {

		return aErr
	}
	return nil
}

func addNewCustomItemsSQL() error {
	query := "ALTER TABLE " + tableUsers + " "
	var execQuery bool
	//
	for key, val := range customAccountInfo {
		// Check if item exists
		checkRows, err := database.Query("SHOW COLUMNS FROM " + tableUsers + " LIKE '" + key + "';")
		if err != nil {
			return err
		}
		//
		checkRows.Next()
		_, colsErr := checkRows.Columns()
		if colsErr != nil {
			// The item doesn't exist yet...
			fmt.Println("Adding AccountInfoColumn '" + key + "'...")
			query = query + "ADD COLUMN " + key + " " + dataTypes[val.dataType]
			if isSizeDataType(val.dataType) {
				query = query + "(" + strconv.Itoa(val.maxSize) + ")"
			} else if isPrecisionDataType(val.dataType) {
				query = query + "(" + strconv.Itoa(val.maxSize) + ", " + strconv.Itoa(val.precision) + ")"
			}
			// Unique check
			if val.unique {
				query = query + " UNIQUE"
			}
			// Not-null check
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
		// Make new columns
		query = query[0:len(query)-2] + ";"
		_, colsErr := database.Exec(query)
		if colsErr != nil {
			return colsErr
		}
	}

	return nil
}