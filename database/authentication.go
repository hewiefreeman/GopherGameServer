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

	SignUpCallback func(string,map[string]interface{})bool                                       = nil
	LoginCallback  func(string,int,map[string]interface{},map[string]interface{})bool            = nil
	DeleteAccountCallback func(string,int,map[string]interface{},map[string]interface{})bool     = nil
	AccountInfoChangeCallback func(string,int,map[string]interface{},map[string]interface{})bool = nil
	PasswordChangeCallback func(string,int,map[string]interface{},map[string]interface{})bool    = nil
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
//   SIGN A USER UP   ////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// SignUpClient signs up the client from the client API with the SQL features enabled.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to sign a
// client up when using the SQL features.
func SignUpClient(userName string, password string, customCols map[string]interface{}) helpers.GopherError {
	if len(userName) == 0 {
		return helpers.NewError(errorRequiredName, helpers.Error_Auth_Required_Name)
	} else if len(password) == 0 {
		return helpers.NewError(errorRequiredPass, helpers.Error_Auth_Required_Pass)
	} else if checkStringSQLInjection(userName) {
		return helpers.NewError(errorMaliciousChars, helpers.Error_Auth_Malicious_Chars)
	} else if !checkCustomRequirements(customCols, customSignupRequirements) {
		return helpers.NewError(errorIncorrectCols, helpers.Error_Auth_Incorrect_Cols)
	}

	//RUN CALLBACK
	if SignUpCallback != nil && !SignUpCallback(userName, customCols) {
		return helpers.NewError(errorDenied, helpers.Error_Action_Denied)
	}

	//ENCRYPT PASSWORD
	passHash, hashErr := helpers.EncryptString(password, encryptionCost)
	if hashErr != nil {
		return helpers.NewError(hashErr.Error(), helpers.Error_Auth_Encryption)
	}

	var vals []interface{} = []interface{}{}

	//CREATE PART 1 OF QUERY
	queryPart1 := "INSERT INTO " + tableUsers + " (" + usersColumnName + ", " + usersColumnPassword + ", "
	if customCols != nil {
		if customLoginColumn != "" {
			if _, ok := customCols[customLoginColumn]; !ok {
				return helpers.NewError(errorInsufficientCols, helpers.Error_Auth_Insufficient_Cols)
			}
		}
		for key, val := range customCols {
			queryPart1 = queryPart1 + key + ", "
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, []interface{}{val, customAccountInfo[key]})
		}
	} else if customLoginColumn != "" {
		return helpers.NewError(errorInsufficientCols, helpers.Error_Auth_Insufficient_Cols)
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
				return helpers.NewError(errorInsufficientCols, helpers.Error_Auth_Insufficient_Cols)
			}
			//CHECK FOR ENCRYPT
			if dt.encrypt {
				value, valueErr = helpers.EncryptString(value, encryptionCost)
				if valueErr != nil {
					return helpers.NewError(valueErr.Error(), helpers.Error_Auth_Encryption)
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
		return helpers.NewError(insertErr.Error(), helpers.Error_Auth_Query)
	}

	//
	return helpers.NewError("", 0)
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
		return "", 0, "", helpers.NewError(errorRequiredName, helpers.Error_Auth_Required_Name)
	} else if len(password) == 0 {
		return "", 0, "", helpers.NewError(errorRequiredPass, helpers.Error_Auth_Required_Pass)
	} else if checkStringSQLInjection(userName) {
		return "", 0, "", helpers.NewError(errorMaliciousChars, helpers.Error_Auth_Malicious_Chars)
	} else if checkStringSQLInjection(deviceTag) {
		return "", 0, "", helpers.NewError(errorMaliciousChars, helpers.Error_Auth_Malicious_Chars)
	} else if !checkCustomRequirements(customCols, customLoginRequirements) {
		return "", 0, "", helpers.NewError(errorIncorrectCols, helpers.Error_Auth_Incorrect_Cols)
	}

	//FIRST THREE ARE id, password, name IN THAT ORDER
	var vals []interface{} = []interface{}{new(int), new([]byte), new(string)}

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select " + usersColumnID + ", " + usersColumnPassword + ", " + usersColumnName + ", "
	if customCols != nil {
		for key := range customCols {
			selectQuery = selectQuery + key + ", "
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}))
		}
	}
	if len(customLoginColumn) > 0 {
		selectQuery = selectQuery[0:len(selectQuery)-2] + " FROM " + tableUsers + " WHERE " + customLoginColumn + "=\"" + userName + "\" LIMIT 1;"
	} else {
		selectQuery = selectQuery[0:len(selectQuery)-2] + " FROM " + tableUsers + " WHERE " + usersColumnName + "=\"" + userName + "\" LIMIT 1;"
	}

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery)
	if err != nil {
		return "", 0, "", helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}
	//
	checkRows.Next()
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close()
		return "", 0, "", helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}
	checkRows.Close()

	//
	dbIndex := vals[0].(*int) // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := vals[1].(*[]byte)
	uName := vals[2].(*string)

	//RUN CALLBACK
	if LoginCallback != nil {
		// GET THE RECIEVED COLUMN VALUES AS MAP
		var recievedVals map[string]interface{} = make(map[string]interface{})
		if customCols != nil {
			i := 3;
			for key := range customCols {
				recievedVals[key] = *(vals[i].(*interface{}))
				//
				i++
			}
		}

		if !LoginCallback(*uName, *dbIndex, recievedVals, customCols) {
			return "", 0, "", helpers.NewError(errorDenied, helpers.Error_Action_Denied)
		}
	}

	//COMPARE HASHED PASSWORDS
	if !helpers.CompareEncryptedData(password, *dbPass) {
		return "", 0, "", helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}

	//AUTO-LOGGING
	var devicePass string
	var devicePassErr error

	if rememberMe && remMe {
		//MAKE AUTO-LOG ENTRY
		devicePass, devicePassErr = helpers.GenerateSecureString(32)
		if devicePassErr == nil {
			database.Exec("INSERT INTO " + tableAutologs + " (" + autologsColumnID + ", " + autologsColumnDeviceTag + ", " + autologsColumnDevicePass +
				") VALUES (" + strconv.Itoa(*dbIndex) + ", \"" + deviceTag + "\", \"" + devicePass + "\");")
		}
	}

	//
	return *uName, *dbIndex, devicePass, helpers.NewError("", 0)
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
		return "", helpers.NewError(errorMaliciousChars, helpers.Error_Auth_Malicious_Chars)
	}
	//EXECUTE SELECT QUERY
	var dPass string
	checkRows, err := database.Query("Select " + autologsColumnDevicePass + " FROM " + tableAutologs + " WHERE " + autologsColumnID + "=" + strconv.Itoa(dbID) + " AND " +
		autologsColumnDeviceTag + "=\"" + tag + "\" LIMIT 1;")
	if err != nil {
		return "", helpers.NewError(errorInvalidAutoLog, helpers.Error_Database_Invalid_Autolog)
	}
	//
	checkRows.Next()
	if scanErr := checkRows.Scan(&dPass); scanErr != nil {
		checkRows.Close()
		return "", helpers.NewError(errorInvalidAutoLog, helpers.Error_Database_Invalid_Autolog)
	}
	checkRows.Close()

	//COMPARE PASSES
	if pass != dPass {
		//SOMEONE TRIED TO COMPROMISE THIS KEY PAIR. DELETE IT NOW.
		RemoveAutoLog(dbID, tag)
	}

	//UPDATE TO NEW PASS
	_, updateErr := database.Exec("UPDATE " + tableAutologs + " SET " + autologsColumnDevicePass + "=\"" + newPass + "\" WHERE " + autologsColumnID + "=" + strconv.Itoa(dbID) + " AND " +
		autologsColumnDeviceTag + "=\"" + tag + "\" LIMIT 1;")
	if updateErr != nil {
		return "", helpers.NewError(errorInvalidAutoLog, helpers.Error_Database_Invalid_Autolog)
	}

	//EVERYTHING WENT WELL, GET THE User's NAME
	var userName string
	userRows, err := database.Query("Select " + usersColumnName + " FROM " + tableUsers + " WHERE " + usersColumnID + "=" + strconv.Itoa(dbID) + " LIMIT 1;")
	if err != nil {
		return "", helpers.NewError(errorInvalidAutoLog, helpers.Error_Database_Invalid_Autolog)
	}
	//
	userRows.Next()
	if scanErr := userRows.Scan(&userName); scanErr != nil {
		userRows.Close()
		return "", helpers.NewError(errorInvalidAutoLog, helpers.Error_Database_Invalid_Autolog)
	}
	userRows.Close()

	//RUN CALLBACK
	if LoginCallback != nil && !LoginCallback(userName, dbID, nil, nil) {
		return "", helpers.NewError(errorDenied, helpers.Error_Action_Denied)
	}

	//
	return userName, helpers.NewError("", 0)
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
		return helpers.NewError(errorRequiredName, helpers.Error_Auth_Required_Name)
	} else if len(password) == 0 {
		return helpers.NewError(errorRequiredPass, helpers.Error_Auth_Required_Pass)
	} else if len(newPassword) == 0 {
		return helpers.NewError(errorRequiredNewPass, helpers.Error_Auth_Required_New_Pass)
	} else if checkStringSQLInjection(userName) {
		return helpers.NewError(errorMaliciousChars, helpers.Error_Auth_Malicious_Chars)
	} else if !checkCustomRequirements(customCols, customPasswordChangeRequirements) {
		return helpers.NewError(errorIncorrectCols, helpers.Error_Auth_Incorrect_Cols)
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(int), new([]byte)}
	var valsList []interface{} = []interface{}{}

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select " + usersColumnID + ", " + usersColumnPassword + ", "
	if customCols != nil {
		for key, val := range customCols {
			selectQuery = selectQuery + key + ", "
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}))
			valsList = append(valsList, []interface{}{val, customAccountInfo[key].dataType, key})
		}
	}
	selectQuery = selectQuery[0:len(selectQuery)-2] + " FROM " + tableUsers + " WHERE " + usersColumnName + "=\"" + userName + "\" LIMIT 1;"

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery)
	if err != nil {
		return helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}
	//
	checkRows.Next()
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close()
		return helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}
	checkRows.Close()

	//
	dbIndex := *(vals[0]).(*int) // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*[]byte)

	//COMPARE HASHED PASSWORDS
	if !helpers.CompareEncryptedData(password, dbPass) {
		return helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}

	//RUN CALLBACK
	if PasswordChangeCallback != nil {
		// GET THE RECIEVED COLUMN VALUES AS MAP
		var recievedVals map[string]interface{} = make(map[string]interface{})
		if customCols != nil {
			i := 2;
			for key := range customCols {
				recievedVals[key] = *(vals[i].(*interface{}))
				//
				i++
			}
		}

		if !PasswordChangeCallback(userName, dbIndex, recievedVals, customCols) {
			return helpers.NewError(errorDenied, helpers.Error_Action_Denied)
		}
	}

	//ENCRYPT NEW PASSWORD
	passHash, hashErr := helpers.EncryptString(newPassword, encryptionCost)
	if hashErr != nil {
		return helpers.NewError(hashErr.Error(), helpers.Error_Auth_Encryption)
	}

	//UPDATE THE PASSWORD
	_, updateErr := database.Exec("UPDATE " + tableUsers + " SET " + usersColumnPassword + "=\"" + passHash + "\" WHERE " + usersColumnID + "=" + strconv.Itoa(dbIndex) + " LIMIT 1;")
	if updateErr != nil {
		return helpers.NewError(updateErr.Error(), helpers.Error_Auth_Query)
	}

	//
	return helpers.NewError("", 0)
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
		return helpers.NewError(errorRequiredName, helpers.Error_Auth_Required_Name)
	} else if len(password) == 0 {
		return helpers.NewError(errorRequiredPass, helpers.Error_Auth_Required_Pass)
	} else if len(customCols) == 0 {
		return helpers.NewError(errorInsufficientCols, helpers.Error_Auth_Insufficient_Cols)
	} else if checkStringSQLInjection(userName) {
		return helpers.NewError(errorMaliciousChars, helpers.Error_Auth_Malicious_Chars)
	} else if !checkCustomRequirements(customCols, customAccountInfoChangeRequirements) {
		return helpers.NewError(errorIncorrectCols, helpers.Error_Auth_Incorrect_Cols)
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(int), new([]byte)}
	var valsList []interface{} = []interface{}{}

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select " + usersColumnID + ", " + usersColumnPassword + ", "
	if customCols != nil {
		for key, val := range customCols {
			selectQuery = selectQuery + key + ", "
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}))
			valsList = append(valsList, []interface{}{val, customAccountInfo[key].dataType, key})
		}
	}
	selectQuery = selectQuery[0:len(selectQuery)-2] + " FROM " + tableUsers + " WHERE " + usersColumnName + "=\"" + userName + "\" LIMIT 1;"

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery)
	if err != nil {
		return helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}
	//
	checkRows.Next()
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close()
		return helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}
	checkRows.Close()

	//
	dbIndex := *(vals[0]).(*int) // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*[]byte)

	//COMPARE HASHED PASSWORDS
	if !helpers.CompareEncryptedData(password, dbPass) {
		return helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}

	//RUN CALLBACK
	if AccountInfoChangeCallback != nil {
		// GET THE RECIEVED COLUMN VALUES AS MAP
		var recievedVals map[string]interface{} = make(map[string]interface{})
		if customCols != nil {
			i := 2;
			for key := range customCols {
				recievedVals[key] = *(vals[i].(*interface{}))
				//
				i++
			}
		}

		if !AccountInfoChangeCallback(userName, dbIndex, recievedVals, customCols) {
			return helpers.NewError(errorDenied, helpers.Error_Action_Denied)
		}
	}

	//MAKE UPDATE QUERY
	updateQuery := "UPDATE " + tableUsers + " SET "
	for i := 0; i < len(valsList); i++ {
		dt := valsList[i].([]interface{})[1].(int)
		//GET STRING VALUE & CHECK FOR INJECTIONS
		value, valueErr := convertDataToString(dataTypes[dt], valsList[i].([]interface{})[0])
		if valueErr != nil {
			return helpers.NewError(valueErr.Error(), helpers.Error_Auth_Conversion)
		}
		//
		updateQuery = updateQuery + valsList[i].([]interface{})[2].(string) + "=" + value + ", "
	}
	updateQuery = updateQuery[0:len(updateQuery)-2] + " WHERE " + usersColumnID + "=" + strconv.Itoa(dbIndex) + " LIMIT 1;"

	//EXECUTE THE UPDATE QUERY
	_, updateErr := database.Exec(updateQuery)
	if updateErr != nil {
		return helpers.NewError(updateErr.Error(), helpers.Error_Auth_Query)
	}

	//
	return helpers.NewError("", 0)
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
		return helpers.NewError(errorRequiredName, helpers.Error_Auth_Required_Name)
	} else if len(password) == 0 {
		return helpers.NewError(errorRequiredPass, helpers.Error_Auth_Required_Pass)
	} else if checkStringSQLInjection(userName) {
		return helpers.NewError(errorMaliciousChars, helpers.Error_Auth_Malicious_Chars)
	} else if !checkCustomRequirements(customCols, customDeleteAccountRequirements) {
		return helpers.NewError(errorIncorrectCols, helpers.Error_Auth_Incorrect_Cols)
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(int), new([]byte)}

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select " + usersColumnID + ", " + usersColumnPassword + ", "
	if customCols != nil {
		for key := range customCols {
			selectQuery = selectQuery + key + ", "
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}))
		}
	}
	selectQuery = selectQuery[0:len(selectQuery)-2] + " FROM " + tableUsers + " WHERE " + usersColumnName + "=\"" + userName + "\" LIMIT 1;"

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery)
	if err != nil {
		return helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}
	//
	checkRows.Next()
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close()
		return helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}
	checkRows.Close()

	//
	dbIndex := *(vals[0]).(*int) // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*[]byte)

	//COMPARE HASHED PASSWORDS
	if !helpers.CompareEncryptedData(password, dbPass) {
		return helpers.NewError(errorIncorrectLogin, helpers.Error_Auth_Incorrect_Login)
	}

	//RUN CALLBACK
	if DeleteAccountCallback != nil {
		// GET THE RECIEVED COLUMN VALUES AS MAP
		var recievedVals map[string]interface{} = make(map[string]interface{})
		if customCols != nil {
			i := 2;
			for key := range customCols {
				recievedVals[key] = *(vals[i].(*interface{}))
				//
				i++
			}
		}

		if !DeleteAccountCallback(userName, dbIndex, recievedVals, customCols) {
			return helpers.NewError(errorDenied, helpers.Error_Action_Denied)
		}
	}

	//REMOVE INSTANCES FROM friends TABLE
	database.Exec("DELETE FROM " + tableFriends + " WHERE " + friendsColumnUser + "=" + strconv.Itoa(dbIndex) + " OR " + friendsColumnFriend + "=" + strconv.Itoa(dbIndex) + ";")

	//DELETE THE ACCOUNT
	_, deleteErr := database.Exec("DELETE FROM " + tableUsers + " WHERE " + usersColumnID + "=" + strconv.Itoa(dbIndex) + " LIMIT 1;")
	if deleteErr != nil {
		return helpers.NewError(deleteErr.Error(), helpers.Error_Auth_Query)
	}

	//
	return helpers.NewError("", 0)
}
