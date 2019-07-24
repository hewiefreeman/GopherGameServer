package database

import (
	"errors"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"strconv"
)

var (
	encryptionCost    int = 4
	customLoginColumn string

	customLoginRequirements             map[string]struct{} = make(map[string]struct{})
	customSignupRequirements            map[string]struct{} = make(map[string]struct{})
	customPasswordChangeRequirements    map[string]struct{} = make(map[string]struct{})
	customAccountInfoChangeRequirements map[string]struct{} = make(map[string]struct{})
	customDeleteAccountRequirements     map[string]struct{} = make(map[string]struct{})

	// SignUpCallback is only for internal Gopher Game Server mechanics.
	SignUpCallback func(string, map[string]interface{}) bool
	// LoginCallback is only for internal Gopher Game Server mechanics.
	LoginCallback func(string, int, map[string]interface{}, map[string]interface{}) bool
	// DeleteAccountCallback is only for internal Gopher Game Server mechanics.
	DeleteAccountCallback func(string, int, map[string]interface{}, map[string]interface{}) bool
	// AccountInfoChangeCallback is only for internal Gopher Game Server mechanics.
	AccountInfoChangeCallback func(string, int, map[string]interface{}, map[string]interface{}) bool
	// PasswordChangeCallback is only for internal Gopher Game Server mechanics.
	PasswordChangeCallback func(string, int, map[string]interface{}, map[string]interface{}) bool
)

// Authentication error messages
const (
	errorDenied           = "Action was denied"
	errorRequiredName     = "A user name is required"
	errorRequiredPass     = "A password is required"
	errorRequiredNewPass  = "A new password is required"
	errorMaliciousChars   = "Malicious characters detected"
	errorIncorrectCols    = "Incorrect custom column data"
	errorInsufficientCols = "Insufficient custom column data"
	errorIncorrectLogin   = "Incorrect login or password"
	errorInvalidAutoLog   = "Invalid auto-login data"
	errorNoShardFound     = "Could not find a master or healthy replica database"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CUSTOM REQUIREMENTS   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// SetCustomSignupRequirements sets the required AccountInfoColumn names for processing a sign up request from a client. If a client
// doesn't send the required info, an error will be sent back.
func SetCustomSignupRequirements(columnNames ...string) error {
	if serverStarted {
		return errors.New("You can't run SetCustomSignupRequirements after the server has started")
	}
	for i := 0; i < len(columnNames); i++ {
		if checkStringSQLInjection(columnNames[i]) {
			return errors.New("Malicious characters detected")
		}
		if _, ok := customAccountInfo[columnNames[i]]; !ok {
			return errors.New("Incorrect column name '" + columnNames[i] + "'")
		}
		customSignupRequirements[columnNames[i]] = struct{}{}
	}
	return nil
}

// SetCustomLoginRequirements sets the required AccountInfoColumn names for processing a login request from a client. If a client
// doesn't send the required info, an error will be sent back.
func SetCustomLoginRequirements(columnNames ...string) error {
	if serverStarted {
		return errors.New("You can't run SetCustomLoginRequirements after the server has started")
	}
	for i := 0; i < len(columnNames); i++ {
		if checkStringSQLInjection(columnNames[i]) {
			return errors.New("Malicious characters detected")
		}
		if _, ok := customAccountInfo[columnNames[i]]; !ok {
			return errors.New("Incorrect column name '" + columnNames[i] + "'")
		}
		customLoginRequirements[columnNames[i]] = struct{}{}
	}
	return nil
}

// SetCustomPasswordChangeRequirements sets the required AccountInfoColumn names for processing a password change request from a client. If a client
// doesn't send the required info, an error will be sent back.
func SetCustomPasswordChangeRequirements(columnNames ...string) error {
	if serverStarted {
		return errors.New("You can't run SetCustomPasswordChangeRequirements after the server has started")
	}
	for i := 0; i < len(columnNames); i++ {
		if checkStringSQLInjection(columnNames[i]) {
			return errors.New("Malicious characters detected")
		}
		if _, ok := customAccountInfo[columnNames[i]]; !ok {
			return errors.New("Incorrect column name '" + columnNames[i] + "'")
		}
		customPasswordChangeRequirements[columnNames[i]] = struct{}{}
	}
	return nil
}

// SetCustomAccountInfoChangeRequirements sets the required AccountInfoColumn names for processing an AccountInfoColumn change request from a client. If a client
// doesn't send the required info, an error will be sent back.
func SetCustomAccountInfoChangeRequirements(columnNames ...string) error {
	if serverStarted {
		return errors.New("You can't run SetCustomAccountInfoChangeRequirements after the server has started")
	}
	for i := 0; i < len(columnNames); i++ {
		if checkStringSQLInjection(columnNames[i]) {
			return errors.New("Malicious characters detected")
		}
		if _, ok := customAccountInfo[columnNames[i]]; !ok {
			return errors.New("Incorrect column name '" + columnNames[i] + "'")
		}
		customAccountInfoChangeRequirements[columnNames[i]] = struct{}{}
	}
	return nil
}

// SetCustomDeleteAccountRequirements sets the required AccountInfoColumn names for processing a delete account request from a client. If a client
// doesn't send the required info, an error will be sent back.
func SetCustomDeleteAccountRequirements(columnNames ...string) error {
	if serverStarted {
		return errors.New("You can't run SetCustomDeleteAccountRequirements after the server has started")
	}
	for i := 0; i < len(columnNames); i++ {
		if checkStringSQLInjection(columnNames[i]) {
			return errors.New("Malicious characters detected")
		}
		if _, ok := customAccountInfo[columnNames[i]]; !ok {
			return errors.New("Incorrect column name '" + columnNames[i] + "'")
		}
		customDeleteAccountRequirements[columnNames[i]] = struct{}{}
	}
	return nil
}

func checkCustomRequirements(customCols map[string]interface{}, requirements map[string]struct{}) bool {
	if customCols != nil && len(requirements) != 0 {
		if len(customCols) == 0 {
			return false
		}
		for key := range customCols {
			if _, ok := requirements[key]; !ok {
				return false
			}
		}
	} else if len(requirements) != 0 {
		return false
	}
	return true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   QUERY HELPERS   /////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SIGN A USER UP   ////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// SignUpClient signs up the client from the client API with the SQL features enabled.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to sign a
// client up when using the SQL features.
func SignUpClient(userName string, password string, customCols map[string]interface{}) helpers.GopherError {
	if len(userName) == 0 {
		return helpers.NewError(errorRequiredName, helpers.ErrorAuthRequiredName)
	} else if len(password) == 0 {
		return helpers.NewError(errorRequiredPass, helpers.ErrorAuthRequiredPass)
	} else if checkStringSQLInjection(userName) {
		return helpers.NewError(errorMaliciousChars, helpers.ErrorAuthMaliciousChars)
	} else if !checkCustomRequirements(customCols, customSignupRequirements) {
		return helpers.NewError(errorIncorrectCols, helpers.ErrorAuthIncorrectCols)
	}

	//RUN CALLBACK
	if SignUpCallback != nil && !SignUpCallback(userName, customCols) {
		return helpers.NewError(errorDenied, helpers.ErrorActionDenied)
	}

	//ENCRYPT PASSWORD
	passHash, hashErr := helpers.EncryptString(password, encryptionCost)
	if hashErr != nil {
		return helpers.NewError(hashErr.Error(), helpers.ErrorAuthEncryption)
	}

	var vals []interface{}

	//CREATE PART 1 OF QUERY
	queryPart1 := "INSERT INTO " + tableUsers + " (" + usersColumnName + ", " + usersColumnPassword + ", "

	if customCols != nil {
		vals = make([]interface{}, 0, len(customCols))
		if customLoginColumn != "" {
			if _, ok := customCols[customLoginColumn]; !ok {
				return helpers.NewError(errorInsufficientCols, helpers.ErrorAuthInsufficientCols)
			}
		}
		for key, val := range customCols {
			queryPart1 = queryPart1 + key + ", "
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, []interface{}{val, customAccountInfo[key]})
		}
	} else if customLoginColumn != "" {
		return helpers.NewError(errorInsufficientCols, helpers.ErrorAuthInsufficientCols)
	}
	queryPart1 = queryPart1[0:len(queryPart1)-2] + ") "

	//CREATE PART 2 OF QUERY
	queryPart2 := "VALUES (\"" + userName + "\", \"" + passHash + "\", "
	if customCols != nil {
		for i := 0; i < len(vals); i++ {
			dt := vals[i].([]interface{})[1].(AccountInfoColumn)
			//GET STRING VALUE & CHECK FOR INJECTIONS
			value, valueErr := convertDataToString(dataTypes[dt.dataType], dt.dataType)
			if valueErr != nil {
				return helpers.NewError(errorInsufficientCols, helpers.ErrorAuthInsufficientCols)
			}
			//CHECK FOR ENCRYPT
			if dt.encrypt {
				value, valueErr = helpers.EncryptString(value, encryptionCost)
				if valueErr != nil {
					return helpers.NewError(valueErr.Error(), helpers.ErrorAuthEncryption)
				}
			}
			//
			queryPart2 = queryPart2 + value + ", "
		}
	}

	queryPart2 = queryPart2[0:len(queryPart2)-2] + ");"

	//EXECUTE QUERY
	_, insertErr := database.Exec(queryPart1 + queryPart2)
	if insertErr != nil {
		return helpers.NewError(insertErr.Error(), helpers.ErrorAuthQuery)
	}

	//
	return helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOGIN CLIENT   //////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// LoginClient logs in the client from the client API with the SQL features enabled.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to log in a
// client when using the SQL features.
func LoginClient(userName string, password string, deviceTag string, remMe bool, customCols map[string]interface{}) (string, int, string, helpers.GopherError) {
	if len(userName) == 0 {
		return "", 0, "", helpers.NewError(errorRequiredName, helpers.ErrorAuthRequiredName)
	} else if len(password) == 0 {
		return "", 0, "", helpers.NewError(errorRequiredPass, helpers.ErrorAuthRequiredPass)
	} else if checkStringSQLInjection(userName) {
		return "", 0, "", helpers.NewError(errorMaliciousChars, helpers.ErrorAuthMaliciousChars)
	} else if checkStringSQLInjection(deviceTag) {
		return "", 0, "", helpers.NewError(errorMaliciousChars, helpers.ErrorAuthMaliciousChars)
	} else if !checkCustomRequirements(customCols, customLoginRequirements) {
		return "", 0, "", helpers.NewError(errorIncorrectCols, helpers.ErrorAuthIncorrectCols)
	}

	//FIRST THREE ARE id, password, name IN THAT ORDER
	var vals []interface{}

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select " + usersColumnID + ", " + usersColumnPassword + ", " + usersColumnName + ", "

	if customCols != nil {
		vals = make([]interface{}, 0, len(customCols)+3)
		vals = append(vals, new(int), new([]byte), new(string))
		for key := range customCols {
			selectQuery = selectQuery + key + ", "
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}))
		}
	} else {
		vals = make([]interface{}, 0, 3)
		vals = append(vals, new(int), new([]byte), new(string))
	}

	// GET LOGIN COLUMN AND TABLE NAME
	var loginCol string = usersColumnName
	var tableName string = tableUsers
	if len(customLoginColumn) > 0 {
		loginCol = customLoginColumn
	}

	selectQuery = selectQuery[0:len(selectQuery)-2] + " FROM " + tableName + " WHERE " + loginCol + "=\"" + userName + "\" LIMIT 1;"

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery)
	if err != nil {
		return "", 0, "", helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}
	//
	checkRows.Next()
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close()
		return "", 0, "", helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}
	checkRows.Close()

	//
	dbIndex := vals[0].(*int) // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := vals[1].(*[]byte)
	uName := vals[2].(*string)

	//RUN CALLBACK
	if LoginCallback != nil {
		// GET THE RECEIVED COLUMN VALUES AS MAP
		var receivedVals map[string]interface{} = make(map[string]interface{})
		if customCols != nil {
			i := 3
			for key := range customCols {
				receivedVals[key] = *(vals[i].(*interface{}))
				//
				i++
			}
		}

		if !LoginCallback(*uName, *dbIndex, receivedVals, customCols) {
			return "", 0, "", helpers.NewError(errorDenied, helpers.ErrorActionDenied)
		}
	}

	//COMPARE HASHED PASSWORDS
	if !helpers.CompareEncryptedData(password, *dbPass) {
		return "", 0, "", helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}

	//AUTO-LOGGING
	var devicePass string
	var devicePassErr error

	if rememberMe && remMe {
		//MAKE AUTO-LOG ENTRY
		devicePass, devicePassErr = helpers.GenerateSecureString(32)
		if devicePassErr == nil {
			_, exErr := database.Exec("INSERT INTO " + tableAutologs + " (" + autologsColumnID + ", " + autologsColumnDeviceTag + ", " + autologsColumnDevicePass +
				") VALUES (" + strconv.Itoa(*dbIndex) + ", \"" + deviceTag + "\", \"" + devicePass + "\");")
			if exErr != nil {
				////// LOG ERROR!!!!!!
			}
		} else {
			////// LOG ERROR!!!!!!
		}
	}

	//
	return *uName, *dbIndex, devicePass, helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   AUTOLOGIN CLIENT   //////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// AutoLoginClient logs in the client automatically with the client API when the SQL features and RememberMe enabled.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to automatically
// log in a client when using the "Remember Me" SQL feature.
func AutoLoginClient(tag string, pass string, newPass string, dbID int) (string, helpers.GopherError) {
	if checkStringSQLInjection(tag) {
		return "", helpers.NewError(errorMaliciousChars, helpers.ErrorAuthMaliciousChars)
	}

	//EXECUTE SELECT QUERY
	var dPass string
	db := database
	tableName := tableAutologs
	checkRows, checkErr := db.Query("Select " + autologsColumnDevicePass + " FROM " + tableName + " WHERE " + autologsColumnID + "=" + strconv.Itoa(dbID) + " AND " +
		autologsColumnDeviceTag + "=\"" + tag + "\" LIMIT 1;")
	if checkErr != nil {
		return "", helpers.NewError(errorInvalidAutoLog, helpers.ErrorDatabaseInvalidAutolog)
	}
	//
	checkRows.Next()
	if scanErr := checkRows.Scan(&dPass); scanErr != nil {
		checkRows.Close()
		return "", helpers.NewError(errorInvalidAutoLog, helpers.ErrorDatabaseInvalidAutolog)
	}
	checkRows.Close()

	//COMPARE PASSES
	if pass != dPass {
		//SOMEONE TRIED TO COMPROMISE THIS KEY PAIR. DELETE IT NOW.
		RemoveAutoLog(dbID, tag)
		return "", helpers.NewError(errorInvalidAutoLog, helpers.ErrorDatabaseInvalidAutolog)
	}

	//UPDATE TO NEW PASS
	_, updateErr := db.Exec("UPDATE " + tableName + " SET " + autologsColumnDevicePass + "=\"" + newPass + "\" WHERE " + autologsColumnID + "=" + strconv.Itoa(dbID) + " AND " +
		autologsColumnDeviceTag + "=\"" + tag + "\" LIMIT 1;")
	if updateErr != nil {
		return "", helpers.NewError(errorInvalidAutoLog, helpers.ErrorDatabaseInvalidAutolog)
	}

	//EVERYTHING WENT WELL, GET THE User's NAME
	tableName = tableUsers
	var userName string
	userRows, usrErr := db.Query("Select " + usersColumnName + " FROM " + tableName + " WHERE " + usersColumnID + "=" + strconv.Itoa(dbID) + " LIMIT 1;")
	if usrErr != nil {
		return "", helpers.NewError(errorInvalidAutoLog, helpers.ErrorDatabaseInvalidAutolog)
	}
	//
	userRows.Next()
	if scanErr := userRows.Scan(&userName); scanErr != nil {
		userRows.Close()
		return "", helpers.NewError(errorInvalidAutoLog, helpers.ErrorDatabaseInvalidAutolog)
	}
	userRows.Close()

	//RUN CALLBACK
	if LoginCallback != nil && !LoginCallback(userName, dbID, nil, nil) {
		return "", helpers.NewError(errorDenied, helpers.ErrorActionDenied)
	}

	//
	return userName, helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   REMOVE AUTOLOGIN KEY PAIR   /////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// RemoveAutoLog removes an auto-log entry when the client API logs out with the SQL features and RememberMe enabled.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. Auto-login entries in the database
// are automatically deleted when a User logs off, or are thought to be compromised by the server.
func RemoveAutoLog(userID int, deviceTag string) {
	if checkStringSQLInjection(deviceTag) {
		return
	}
	database.Exec("DELETE FROM " + tableAutologs + " WHERE " + autologsColumnID + "=" + strconv.Itoa(userID) + " AND " + autologsColumnDeviceTag + "=\"" + deviceTag + "\" LIMIT 1;")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHANGE PASSWORD   ///////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// ChangePassword changes a client's password with the SQL features enabled.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to change
// a user's password when using the SQL features.
func ChangePassword(userName string, password string, newPassword string, customCols map[string]interface{}) helpers.GopherError {
	if len(userName) == 0 {
		return helpers.NewError(errorRequiredName, helpers.ErrorAuthRequiredName)
	} else if len(password) == 0 {
		return helpers.NewError(errorRequiredPass, helpers.ErrorAuthRequiredPass)
	} else if len(newPassword) == 0 {
		return helpers.NewError(errorRequiredNewPass, helpers.ErrorAuthRequiredNewPass)
	} else if checkStringSQLInjection(userName) {
		return helpers.NewError(errorMaliciousChars, helpers.ErrorAuthMaliciousChars)
	} else if !checkCustomRequirements(customCols, customPasswordChangeRequirements) {
		return helpers.NewError(errorIncorrectCols, helpers.ErrorAuthIncorrectCols)
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{}
	var valsList []interface{}

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select " + usersColumnID + ", " + usersColumnPassword + ", "
	if customCols != nil {
		vals = make([]interface{}, 0, len(customCols)+2)
		vals = append(vals, new(int), new([]byte))
		valsList = make([]interface{}, 0, len(customCols))
		for key, val := range customCols {
			selectQuery = selectQuery + key + ", "
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}))
			valsList = append(valsList, []interface{}{val, customAccountInfo[key].dataType, key})
		}
	} else {
		vals = make([]interface{}, 0, 2)
		vals = append(vals, new(int), new([]byte))
	}
	selectQuery = selectQuery[0:len(selectQuery)-2] + " FROM " + tableUsers + " WHERE " + usersColumnName + "=\"" + userName + "\" LIMIT 1;"

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery)
	if err != nil {
		return helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}
	//
	checkRows.Next()
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close()
		return helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}
	checkRows.Close()

	//
	dbIndex := *(vals[0]).(*int) // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*[]byte)

	//COMPARE HASHED PASSWORDS
	if !helpers.CompareEncryptedData(password, dbPass) {
		return helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}

	//RUN CALLBACK
	if PasswordChangeCallback != nil {
		// GET THE RECEIVED COLUMN VALUES AS MAP
		var receivedVals map[string]interface{} = make(map[string]interface{})
		if customCols != nil {
			i := 2
			for key := range customCols {
				receivedVals[key] = *(vals[i].(*interface{}))
				//
				i++
			}
		}

		if !PasswordChangeCallback(userName, dbIndex, receivedVals, customCols) {
			return helpers.NewError(errorDenied, helpers.ErrorActionDenied)
		}
	}

	//ENCRYPT NEW PASSWORD
	passHash, hashErr := helpers.EncryptString(newPassword, encryptionCost)
	if hashErr != nil {
		return helpers.NewError(hashErr.Error(), helpers.ErrorAuthEncryption)
	}

	//UPDATE THE PASSWORD
	_, updateErr := database.Exec("UPDATE " + tableUsers + " SET " + usersColumnPassword + "=\"" + passHash + "\" WHERE " + usersColumnID + "=" + strconv.Itoa(dbIndex) + " LIMIT 1;")
	if updateErr != nil {
		return helpers.NewError(updateErr.Error(), helpers.ErrorAuthQuery)
	}

	//
	return helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHANGE ACCOUNT INFO   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// ChangeAccountInfo changes a client's AccountInfoColumn with the SQL features enabled.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to change
// a user's AccountInfoColumn when using the SQL features.
func ChangeAccountInfo(userName string, password string, customCols map[string]interface{}) helpers.GopherError {
	if len(userName) == 0 {
		return helpers.NewError(errorRequiredName, helpers.ErrorAuthRequiredName)
	} else if len(password) == 0 {
		return helpers.NewError(errorRequiredPass, helpers.ErrorAuthRequiredPass)
	} else if len(customCols) == 0 {
		return helpers.NewError(errorInsufficientCols, helpers.ErrorAuthInsufficientCols)
	} else if checkStringSQLInjection(userName) {
		return helpers.NewError(errorMaliciousChars, helpers.ErrorAuthMaliciousChars)
	} else if !checkCustomRequirements(customCols, customAccountInfoChangeRequirements) {
		return helpers.NewError(errorIncorrectCols, helpers.ErrorAuthIncorrectCols)
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{}
	var valsList []interface{}

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select " + usersColumnID + ", " + usersColumnPassword + ", "
	if customCols != nil {
		vals = make([]interface{}, 0, len(customCols)+2)
		vals = append(vals, new(int), new([]byte))
		valsList = make([]interface{}, 0, len(customCols))
		for key, val := range customCols {
			selectQuery = selectQuery + key + ", "
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}))
			valsList = append(valsList, []interface{}{val, customAccountInfo[key].dataType, key})
		}
	} else {
		vals = make([]interface{}, 0, 2)
		vals = append(vals, new(int), new([]byte))
	}
	selectQuery = selectQuery[0:len(selectQuery)-2] + " FROM " + tableUsers + " WHERE " + usersColumnName + "=\"" + userName + "\" LIMIT 1;"

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery)
	if err != nil {
		return helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}
	//
	checkRows.Next()
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close()
		return helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}
	checkRows.Close()

	//
	dbIndex := *(vals[0]).(*int) // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*[]byte)

	//COMPARE HASHED PASSWORDS
	if !helpers.CompareEncryptedData(password, dbPass) {
		return helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}

	//RUN CALLBACK
	if AccountInfoChangeCallback != nil {
		// GET THE RECEIVED COLUMN VALUES AS MAP
		var receivedVals map[string]interface{} = make(map[string]interface{})
		if customCols != nil {
			i := 2
			for key := range customCols {
				receivedVals[key] = *(vals[i].(*interface{}))
				//
				i++
			}
		}

		if !AccountInfoChangeCallback(userName, dbIndex, receivedVals, customCols) {
			return helpers.NewError(errorDenied, helpers.ErrorActionDenied)
		}
	}

	//MAKE UPDATE QUERY
	updateQuery := "UPDATE " + tableUsers + " SET "
	for i := 0; i < len(valsList); i++ {
		dt := valsList[i].([]interface{})[1].(int)
		//GET STRING VALUE & CHECK FOR INJECTIONS
		value, valueErr := convertDataToString(dataTypes[dt], valsList[i].([]interface{})[0])
		if valueErr != nil {
			return helpers.NewError(valueErr.Error(), helpers.ErrorAuthConversion)
		}
		//
		updateQuery = updateQuery + valsList[i].([]interface{})[2].(string) + "=" + value + ", "
	}
	updateQuery = updateQuery[0:len(updateQuery)-2] + " WHERE " + usersColumnID + "=" + strconv.Itoa(dbIndex) + " LIMIT 1;"

	//EXECUTE THE UPDATE QUERY
	_, updateErr := database.Exec(updateQuery)
	if updateErr != nil {
		return helpers.NewError(updateErr.Error(), helpers.ErrorAuthQuery)
	}

	//
	return helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   DELETE CLIENT ACCOUNT   /////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// DeleteAccount deletes a client's account with the SQL features enabled.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to delete a
// user's account when using the SQL features.
func DeleteAccount(userName string, password string, customCols map[string]interface{}) helpers.GopherError {
	if len(userName) == 0 {
		return helpers.NewError(errorRequiredName, helpers.ErrorAuthRequiredName)
	} else if len(password) == 0 {
		return helpers.NewError(errorRequiredPass, helpers.ErrorAuthRequiredPass)
	} else if checkStringSQLInjection(userName) {
		return helpers.NewError(errorMaliciousChars, helpers.ErrorAuthMaliciousChars)
	} else if !checkCustomRequirements(customCols, customDeleteAccountRequirements) {
		return helpers.NewError(errorIncorrectCols, helpers.ErrorAuthIncorrectCols)
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{}

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select " + usersColumnID + ", " + usersColumnPassword + ", "
	if customCols != nil {
		vals = make([]interface{}, 0, len(customCols)+2)
		vals = append(vals, new(int), new([]byte))
		for key := range customCols {
			selectQuery = selectQuery + key + ", "
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}))
		}
	} else {
		vals = make([]interface{}, 0, 2)
		vals = append(vals, new(int), new([]byte))
	}
	selectQuery = selectQuery[0:len(selectQuery)-2] + " FROM " + tableUsers + " WHERE " + usersColumnName + "=\"" + userName + "\" LIMIT 1;"

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery)
	if err != nil {
		return helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}
	//
	checkRows.Next()
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close()
		return helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}
	checkRows.Close()

	//
	dbIndex := *(vals[0]).(*int) // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*[]byte)

	//COMPARE HASHED PASSWORDS
	if !helpers.CompareEncryptedData(password, dbPass) {
		return helpers.NewError(errorIncorrectLogin, helpers.ErrorAuthIncorrectLogin)
	}

	//RUN CALLBACK
	if DeleteAccountCallback != nil {
		// GET THE RECEIVED COLUMN VALUES AS MAP
		var receivedVals map[string]interface{} = make(map[string]interface{})
		if customCols != nil {
			i := 2
			for key := range customCols {
				receivedVals[key] = *(vals[i].(*interface{}))
				//
				i++
			}
		}

		if !DeleteAccountCallback(userName, dbIndex, receivedVals, customCols) {
			return helpers.NewError(errorDenied, helpers.ErrorActionDenied)
		}
	}

	//REMOVE INSTANCES FROM friends TABLE
	database.Exec("DELETE FROM " + tableFriends + " WHERE " + friendsColumnUser + "=" + strconv.Itoa(dbIndex) + " OR " + friendsColumnFriend + "=" + strconv.Itoa(dbIndex) + ";")

	//DELETE THE ACCOUNT
	_, deleteErr := database.Exec("DELETE FROM " + tableUsers + " WHERE " + usersColumnID + "=" + strconv.Itoa(dbIndex) + " LIMIT 1;")
	if deleteErr != nil {
		return helpers.NewError(deleteErr.Error(), helpers.ErrorAuthQuery)
	}

	//
	return helpers.NoError()
}
