package database

import(
	"errors"
	"strconv"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"fmt"
)

var(
	encryptionCost int = 4;
	customLoginColumn string = "";
	customLoginRequirements map[string]struct{} = make(map[string]struct{});
	customSignupRequirements map[string]struct{} = make(map[string]struct{});
	customPasswordChangeRequirements map[string]struct{} = make(map[string]struct{});
	customAccountInfoChangeRequirements map[string]struct{} = make(map[string]struct{});
	customDeleteAccountRequirements map[string]struct{} = make(map[string]struct{});
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CUSTOM REQUIREMENTS   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Sets the required AccountInfoColumn names for processing a sign up request from a client. If a client
// doesn't send the required info, an error will be sent back.
func SetCustomSignupRequirements(columnNames ...string) error {
	if(serverStarted){
		return errors.New("You can't run SetCustomSignupRequirements after the server has started");
	}
	for i := 0; i < len(columnNames); i++ {
		if(checkStringSQLInjection(columnNames[i])){
			return errors.New("Malicious characters detected");
		}
		if _, ok := customAccountInfo[columnNames[i]]; !ok {
			return errors.New("Incorrect column name '"+columnNames[i]+"'");
		}else{
			customSignupRequirements[columnNames[i]] = struct{}{};
		}
	}
	return nil;
}

// Sets the required AccountInfoColumn names for processing a login request from a client. If a client
// doesn't send the required info, an error will be sent back.
func SetCustomLoginRequirements(columnNames ...string) error {
	if(serverStarted){
		return errors.New("You can't run SetCustomLoginRequirements after the server has started");
	}
	for i := 0; i < len(columnNames); i++ {
		if(checkStringSQLInjection(columnNames[i])){
			return errors.New("Malicious characters detected");
		}
		if _, ok := customAccountInfo[columnNames[i]]; !ok {
			return errors.New("Incorrect column name '"+columnNames[i]+"'");
		}else{
			customLoginRequirements[columnNames[i]] = struct{}{};
		}
	}
	return nil;
}

// Sets the required AccountInfoColumn names for processing a password change request from a client. If a client
// doesn't send the required info, an error will be sent back.
func SetCustomPasswordChangeRequirements(columnNames ...string) error {
	if(serverStarted){
		return errors.New("You can't run SetCustomPasswordChangeRequirements after the server has started");
	}
	for i := 0; i < len(columnNames); i++ {
		if(checkStringSQLInjection(columnNames[i])){
			return errors.New("Malicious characters detected");
		}
		if _, ok := customAccountInfo[columnNames[i]]; !ok {
			return errors.New("Incorrect column name '"+columnNames[i]+"'");
		}else{
			customPasswordChangeRequirements[columnNames[i]] = struct{}{};
		}
	}
	return nil;
}

// Sets the required AccountInfoColumn names for processing an AccountInfoColumn change request from a client. If a client
// doesn't send the required info, an error will be sent back.
func SetCustomAccountInfoChangeRequirements(columnNames ...string) error {
	if(serverStarted){
		return errors.New("You can't run SetCustomAccountInfoChangeRequirements after the server has started");
	}
	for i := 0; i < len(columnNames); i++ {
		if(checkStringSQLInjection(columnNames[i])){
			return errors.New("Malicious characters detected");
		}
		if _, ok := customAccountInfo[columnNames[i]]; !ok {
			return errors.New("Incorrect column name '"+columnNames[i]+"'");
		}else{
			customAccountInfoChangeRequirements[columnNames[i]] = struct{}{};
		}
	}
	return nil;
}

// Sets the required AccountInfoColumn names for processing a delete account request from a client. If a client
// doesn't send the required info, an error will be sent back.
func SetCustomDeleteAccountRequirements(columnNames ...string) error {
	if(serverStarted){
		return errors.New("You can't run SetCustomDeleteAccountRequirements after the server has started");
	}
	for i := 0; i < len(columnNames); i++ {
		if(checkStringSQLInjection(columnNames[i])){
			return errors.New("Malicious characters detected");
		}
		if _, ok := customAccountInfo[columnNames[i]]; !ok {
			return errors.New("Incorrect column name '"+columnNames[i]+"'");
		}else{
			customDeleteAccountRequirements[columnNames[i]] = struct{}{};
		}
	}
	return nil;
}

func checkCustomRequirements(customCols map[string]interface{}, requirements map[string]struct{}) bool {
	if(customCols != nil && len(requirements) != 0){
		if(len(customCols) == 0){
			return false;
		}
		for key, _ := range customCols {
			if _, ok := requirements[key]; !ok {
				return false;
			}
		}
	}else if(len(requirements) != 0){
		return false;
	}
	return true;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SIGN A USER UP   ////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to sign a
// client up when using the SQL features.
func SignUpClient(userName string, password string, customCols map[string]interface{}) error {
	if(len(userName) == 0){
		return errors.New("A user name is required to sign up");
	}else if(len(password) == 0){
		return errors.New("A password is required to sign up");
	}else if(checkStringSQLInjection(userName)){
		return errors.New("Malicious characters detected");
	}else if(!checkCustomRequirements(customCols, customSignupRequirements)){
		return errors.New("Incorrect data supplied");
	}

	//ENCRYPT PASSWORD
	passHash, hashErr := helpers.HashPassword(password, encryptionCost);
	if(hashErr != nil){ return hashErr; }

	var vals []interface{} = []interface{}{};

	//CREATE PART 1 OF QUERY
	queryPart1 := "INSERT INTO "+tableUsers+" ("+usersColumnName+", "+usersColumnPassword+", ";
	if(customCols != nil){
		if(customLoginColumn != ""){
			if _, ok := customCols[customLoginColumn]; !ok {
				return errors.New("Insufficient data supplied");
			}
		}
		for key, val := range customCols {
			queryPart1 = queryPart1+key+", ";
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, []interface{}{val, customAccountInfo[key].dataType});
		}
	}else if(customLoginColumn != ""){
		return errors.New("Insufficient data supplied");
	}
	queryPart1 = queryPart1[0:len(queryPart1)-2]+") ";

	//CREATE PART 2 OF QUERY
	queryPart2 := "VALUES (\""+userName+"\", \""+passHash+"\", ";
	if(customCols != nil){
		for i := 0; i < len(vals); i++ {
			dt := vals[i].([]interface{})[1].(int);
			//GET STRING VALUE & CHECK FOR INJECTIONS
			value, valueErr := convertDataToString(dataTypes[dt], vals[i].([]interface{})[0]);
			if(valueErr != nil){ return valueErr; }
			//
			queryPart2 = queryPart2+value+", ";
		}
	}

	queryPart2 = queryPart2[0:len(queryPart2)-2]+");";

	//EXECUTE QUERY
	_, insertErr := database.Exec(queryPart1+queryPart2);
	if(insertErr != nil){ return insertErr; }

	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOGIN CLIENT   //////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to log in a
// client when using the SQL features.
func LoginClient(userName string, password string, deviceTag string, remMe bool, customCols map[string]interface{}) (string, int, string, error) {
	if(len(userName) == 0){
		return "", 0, "", errors.New("A user name is required to log in");
	}else if(len(password) == 0){
		return "", 0, "", errors.New("A password is required to log in");
	}else if(checkStringSQLInjection(userName)){
		return "", 0, "", errors.New("Malicious characters detected");
	}else if(checkStringSQLInjection(deviceTag)){
		return "", 0, "", errors.New("Malicious characters detected");
	}else if(!checkCustomRequirements(customCols, customLoginRequirements)){
		return "", 0, "", errors.New("Incorrect data supplied");
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(int), new([]byte), new(string)};

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select "+usersColumnID+", "+usersColumnPassword+", "+usersColumnName+", ";
	if(customCols != nil){
		for key, _ := range customCols {
			selectQuery = selectQuery+key+", ";
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}));
		}
	}
	if(len(customLoginColumn) > 0){
		selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+customLoginColumn+"=\""+userName+"\" LIMIT 1;";
	}else{
		selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+usersColumnName+"=\""+userName+"\" LIMIT 1;";
	}


	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery);
	if(err != nil){ return "", 0, "", err; }
	//
	checkRows.Next();
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close();
		return "", 0, "", errors.New("Login or password is incorrect");
	}
	checkRows.Close();

	//
	dbIndex := vals[0].(*int); // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := vals[1].(*[]byte);
	uName := vals[2].(*string);

	//COMPARE HASHED PASSWORDS
	if(!helpers.CheckPasswordHash(password, *dbPass)){
		return "", 0, "", errors.New("Login or password is incorrect");
	}

	//AUTO-LOGGING
	var devicePass string = "";
	var devicePassErr error = nil;

	if(rememberMe && remMe){
		//MAKE AUTO-LOG ENTRY
		devicePass, devicePassErr = helpers.GenerateSecureString(32);
		if(devicePassErr == nil){
			database.Exec("INSERT INTO "+tableAutologs+" ("+autologsColumnID+", "+autologsColumnDeviceTag+", "+autologsColumnDevicePass+
								") VALUES ("+strconv.Itoa(*dbIndex)+", \""+deviceTag+"\", \""+devicePass+"\");");
		}
	}

	//
	return *uName, *dbIndex, devicePass, nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   AUTOLOGIN CLIENT   //////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to automatically
// log in a client when using the "Remember Me" SQL feature.
func AutoLoginClient(tag string, pass string, newPass string, dbID int) (string, error) {
	if(checkStringSQLInjection(tag)){
		return "", errors.New("Malicious characters detected");
	}
	//EXECUTE SELECT QUERY
	var dPass string;
	checkRows, err := database.Query("Select "+autologsColumnDevicePass+" FROM "+tableAutologs+" WHERE "+autologsColumnID+"="+strconv.Itoa(dbID)+" AND "+
								autologsColumnDeviceTag+"=\""+tag+"\" LIMIT 1;");
	if(err != nil){ return "", errors.New("Incorrect autolog info"); }
	//
	checkRows.Next();
	if scanErr := checkRows.Scan(&dPass); scanErr != nil {
		checkRows.Close();
		return "", errors.New("Incorrect autolog info");
	}
	checkRows.Close();

	//COMPARE PASSES
	if(pass != dPass){
		//SOMEONE TRIED TO COMPROMISE THIS KEY PAIR. DELETE IT NOW.
		RemoveAutoLog(dbID, tag);
	}

	//UPDATE TO NEW PASS
	_, updateErr := database.Exec("UPDATE "+tableAutologs+" SET "+autologsColumnDevicePass+"=\""+newPass+"\" WHERE "+autologsColumnID+"="+strconv.Itoa(dbID)+" AND "+
							autologsColumnDeviceTag+"=\""+tag+"\" LIMIT 1;");
	if(updateErr != nil){ return "", errors.New("Incorrect autolog info"); }

	//EVERYTHING WENT WELL, GET THE User's NAME
	var userName string;
	userRows, err := database.Query("Select "+usersColumnName+" FROM "+tableUsers+" WHERE "+usersColumnID+"="+strconv.Itoa(dbID)+" LIMIT 1;");
	if(err != nil){ return "", errors.New("Incorrect autolog info"); }
	//
	userRows.Next();
	if scanErr := userRows.Scan(&userName); scanErr != nil {
		userRows.Close();
		return "", errors.New("Incorrect autolog info");
	}
	userRows.Close();

	//
	return userName, nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   REMOVE AUTOLOGIN KEY PAIR   /////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Auto-login entries in the database
// are automatically deleted when a User logs off, or are thought to be compromised by the server.
func RemoveAutoLog(userID int, deviceTag string){
	if(checkStringSQLInjection(deviceTag)){
		return;
	}
	database.Exec("DELETE FROM "+tableAutologs+" WHERE "+autologsColumnID+"="+strconv.Itoa(userID)+" AND "+autologsColumnDeviceTag+"=\""+deviceTag+"\" LIMIT 1;");
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHANGE PASSWORD   ///////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to change
// a user's password when using the SQL features.
func ChangePassword(userName string, password string, newPassword string, customCols map[string]interface{}) error {
	if(len(userName) == 0){
		return errors.New("A user name is required to change a password");
	}else if(len(password) == 0){
		return errors.New("A password is required to change a password");
	}else if(len(newPassword) == 0){
		return errors.New("A new password is required to change a password");
	}else if(checkStringSQLInjection(userName)){
		return errors.New("Malicious characters detected");
	}else if(!checkCustomRequirements(customCols, customPasswordChangeRequirements)){
		return errors.New("Incorrect data supplied");
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(int), new([]byte)};
	var valsList []interface{} = []interface{}{};

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select "+usersColumnID+", "+usersColumnPassword+", ";
	if(customCols != nil){
		for key, val := range customCols {
			selectQuery = selectQuery+key+", ";
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}));
			valsList = append(valsList, []interface{}{val, customAccountInfo[key].dataType, key});
		}
	}
	selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+usersColumnName+"=\""+userName+"\" LIMIT 1;";

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery);
	if(err != nil){ return err; }
	//
	checkRows.Next();
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close();
		return errors.New("Login or password is incorrect");
	}
	checkRows.Close();

	//
	dbIndex := *(vals[0]).(*int); // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*[]byte);

	//COMPARE HASHED PASSWORDS
	if(!helpers.CheckPasswordHash(password, dbPass)){
		return errors.New("Login or password is incorrect");
	}

	//ENCRYPT NEW PASSWORD
	passHash, hashErr := helpers.HashPassword(newPassword, encryptionCost);
	if(hashErr != nil){ return hashErr; }

	//UPDATE THE PASSWORD
	_, updateErr := database.Exec("UPDATE "+tableUsers+" SET "+usersColumnPassword+"=\""+passHash+"\" WHERE "+usersColumnID+"="+strconv.Itoa(dbIndex)+" LIMIT 1;");
	if(updateErr != nil){ return updateErr; }

	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHANGE ACCOUNT INFO   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to change
// a user's AccountInfoColumn when using the SQL features.
func ChangeAccountInfo(userName string, password string, customCols map[string]interface{}) error {
	if(len(userName) == 0){
		return errors.New("A user name is required to change custom account info");
	}else if(len(password) == 0){
		return errors.New("A password is required to change custom account info");
	}else if(len(customCols) == 0){
		return errors.New("Custom columns need to be provided to change their info");
	}else if(checkStringSQLInjection(userName)){
		return errors.New("Malicious characters detected");
	}else if(!checkCustomRequirements(customCols, customAccountInfoChangeRequirements)){
		return errors.New("Incorrect data supplied");
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(int), new([]byte)};
	var valsList []interface{} = []interface{}{};

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select "+usersColumnID+", "+usersColumnPassword+", ";
	if(customCols != nil){
		for key, val := range customCols {
			selectQuery = selectQuery+key+", ";
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}));
			valsList = append(valsList, []interface{}{val, customAccountInfo[key].dataType, key});
		}
	}
	selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+usersColumnName+"=\""+userName+"\" LIMIT 1;";

	fmt.Println(selectQuery);

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery);
	if(err != nil){ return err; }
	//
	checkRows.Next();
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		fmt.Println(scanErr);
		checkRows.Close();
		return errors.New("Login or password is incorrect");
	}
	checkRows.Close();
	fmt.Println("got past scan.");

	//
	dbIndex := *(vals[0]).(*int); // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*[]byte);

	//COMPARE HASHED PASSWORDS
	if(!helpers.CheckPasswordHash(password, dbPass)){
		return errors.New("Login or password is incorrect");
	}
	fmt.Println("got past password check");

	//UPDATE THE AccountInfoColumns

	//MAKE UPDATE QUERY
	updateQuery := "UPDATE "+tableUsers+" SET ";
	for i := 0; i < len(valsList); i++ {
		dt := valsList[i].([]interface{})[1].(int);
		//GET STRING VALUE & CHECK FOR INJECTIONS
		value, valueErr := convertDataToString(dataTypes[dt], valsList[i].([]interface{})[0]);
		if(valueErr != nil){ return valueErr; }
		//
		updateQuery = updateQuery+valsList[i].([]interface{})[2].(string)+"="+value+", ";
	}
	updateQuery = updateQuery[0:len(updateQuery)-2]+" WHERE "+usersColumnID+"="+strconv.Itoa(dbIndex)+" LIMIT 1;";

	//EXECUTE THE UPDATE QUERY
	_, updateErr := database.Exec(updateQuery);
	if(updateErr != nil){ return updateErr; }

	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   DELETE CLIENT ACCOUNT   /////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to delete a
// user's account when using the SQL features.
func DeleteAccount(userName string, password string, customCols map[string]interface{}) error {
	if(len(userName) == 0){
		return errors.New("A user name is required to delete an account");
	}else if(len(password) == 0){
		return errors.New("A password is required to delete an account");
	}else if(checkStringSQLInjection(userName)){
		return errors.New("Malicious characters detected");
	}else if(!checkCustomRequirements(customCols, customDeleteAccountRequirements)){
		return errors.New("Incorrect data supplied");
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(int), new([]byte)};

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select "+usersColumnID+", "+usersColumnPassword+", ";
	if(customCols != nil){
		for key, _ := range customCols {
			selectQuery = selectQuery+key+", ";
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}));
		}
	}
	selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+usersColumnName+"=\""+userName+"\" LIMIT 1;";

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery);
	if(err != nil){ return err; }
	//
	checkRows.Next();
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close();
		return errors.New("Login or password is incorrect");
	}
	checkRows.Close();

	//
	dbIndex := *(vals[0]).(*int); // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*[]byte);

	//COMPARE HASHED PASSWORDS
	if(!helpers.CheckPasswordHash(password, dbPass)){
		return errors.New("Login or password is incorrect");
	}

	//EVERYTHING WENT FINE AND DANDY, DELETE THE ACCOUNT
	_, deleteErr := database.Exec("DELETE FROM "+tableUsers+" WHERE "+usersColumnID+"="+strconv.Itoa(dbIndex)+" LIMIT 1;");
	if(deleteErr != nil){ return deleteErr; }

	//REMOVE INSTANCES FROM friends TABLE
	database.Exec("DELETE FROM "+tableFriends+" WHERE "+friendsColumnUser+"="+strconv.Itoa(dbIndex)+" OR "+friendsColumnFriend+"="+strconv.Itoa(dbIndex)+";");

	//
	return nil;
}
