// Package database contains helpers for customizing your database with the SQL features enabled.
// It mostly contains a bunch of mixed Gopher Server only functions and customizing methods.
// It would probably be easier to take a look at the database usage section on the Github page
// for the project before looking through here for more info.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Github project page specifies to use blank import
	"strconv"
)

var (
	//THE DATABASE
	database *sql.DB

	//SERVER SETTINGS
	serverStarted bool   = false
	serverPaused  bool   = false
	rememberMe    bool   = false
	databaseName  string = "gopherDB"
	inited        bool   = false
)

//TABLE & COLUMN NAMES
const (
	tableUsers    = "users"
	tableFriends  = "friends"
	tableAutologs = "autologs"

	//users TABLE COLUMNS
	usersColumnID       = "_id"
	usersColumnName     = "name"
	usersColumnPassword = "pass"

	//friends TABLE COLUMNS
	friendsColumnUser   = "user"
	friendsColumnFriend = "friend"
	friendsColumnStatus = "status"

	//autologs TABLE COLUMNS
	autologsColumnID         = "_id"
	autologsColumnDeviceTag  = "dn"
	autologsColumnDevicePass = "da"
)

// Init initializes the database connection and sets up the database according to your custom parameters.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. If you want to enable SQL authorization
// and friending, use the EnableSqlFeatures and corresponding options in ServerSetting.
func Init(userName string, password string, dbName string, protocol string, ip string, port int, encryptCost int, remMe bool, custLoginCol string) error {
	if inited {
		return errors.New("sql package is already initialized")
	} else if len(userName) == 0 {
		return errors.New("sql.Start() requires a user name")
	} else if len(password) == 0 {
		return errors.New("sql.Start() requires a password")
	} else if len(userName) == 0 {
		return errors.New("sql.Start() requires a database name")
	} else if len(custLoginCol) > 0 {
		if _, ok := customAccountInfo[custLoginCol]; !ok {
			return errors.New("The AccountInfoColumn '" + custLoginCol + "' does not exist. Use database.NewAccountInfoColumn() to make a column with that name.")
		}
		customLoginColumn = custLoginCol
	}

	if encryptCost >= 4 && encryptCost <= 31 {
		encryptionCost = encryptCost
	} else if encryptCost != 0 {
		fmt.Println("EncryptionCost must be a minimum of 4, and max of 31. Setting to default: 4")
	}

	rememberMe = remMe

	var err error

	//OPEN THE DATABASE
	database, err = sql.Open("mysql", userName+":"+password+"@"+protocol+"("+ip+":"+strconv.Itoa(port)+")/"+dbName)
	if err != nil {
		return err
	}
	//NOTE: Open doesn't open a connection.
	//MUST PING TO CHECK IF FOUND DATABASE
	err = database.Ping()
	if err != nil {
		return errors.New("Could not connect to database!")
	}

	if len(dbName) != 0 {
		databaseName = dbName
	}

	//CONFIGURE DATABASE
	err = setUp()
	if err != nil {
		return err
	}

	//
	inited = true

	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   GET User's DATABASE INDEX   /////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// GetUserDatabaseIndex gets the database index of a User by their name.
func GetUserDatabaseIndex(userName string) (int, error) {
	if checkStringSQLInjection(userName) {
		return 0, errors.New("Malicious characters detected")
	}
	var id int
	rows, err := database.Query("SELECT " + usersColumnID + " FROM " + tableUsers + " WHERE " + usersColumnName + "=\"" + userName + "\" LIMIT 1;")
	if err != nil {
		return 0, err
	}
	//
	rows.Next()
	if scanErr := rows.Scan(&id); scanErr != nil {
		rows.Close()
		return 0, scanErr
	}
	rows.Close()

	//
	return id, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER STARTUP FUNCTIONS   ///////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetServerStarted is for Gopher Game Server internal mechanics only.
func SetServerStarted(val bool) {
	if !serverStarted {
		serverStarted = val
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER PAUSE AND RESUME   ///////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Pause is only for internal Gopher Game Server mechanics.
func Pause() {
	if !serverPaused {
		serverPaused = true
		serverStarted = false
	}
}

// Resume is only for internal Gopher Game Server mechanics.
func Resume() {
	if serverPaused {
		serverStarted = true
		serverPaused = false
	}
}
