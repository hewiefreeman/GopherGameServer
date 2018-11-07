package database

import(
	"errors"
	"strconv"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"fmt"
)

var(
	encryptionCost int = 32;
)

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
	}

	// CHECK FOR VALID COLUMNS IN customCols - ALSO PREVENTS INJECTIONS IN KEYS
	if(customCols != nil){
		for key, _ := range customCols {
			if _, ok := customAccountInfo[key]; !ok {
				return errors.New("Incorrect data supplied!");
			}
		}
	}

	//ENCRYPT PASSWORD
	passHash, hashErr := helpers.HashPassword(password, encryptionCost);
	if(hashErr != nil){ return hashErr; }

	var vals []interface{} = []interface{}{};

	//CREATE PART 1 OF QUERY
	queryPart1 := "INSERT INTO "+tableUsers+" ("+usersColumnName+", "+usersColumnPassword+", ";
	if(customCols != nil){
		for key, val := range customCols {
			queryPart1 = queryPart1+key+", ";
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, []interface{}{val, customAccountInfo[key].dataType});
		}
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
			//CHECK IF DATA TYPE NEEDS QUOTES
			if(dataTypeNeedsQuotes(dt)){
				queryPart2 = queryPart2+"\""+value+"\", ";
			}else{
				queryPart2 = queryPart2+value+", ";
			}
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
func LoginClient(userName string, password string, customCols map[string]interface{}) error {
	if(len(userName) == 0){
		return errors.New("A user name is required to log in");
	}else if(len(password) == 0){
		return errors.New("A password is required to log in");
	}else if(checkStringSQLInjection(userName)){
		return errors.New("Malicious characters detected");
	}

	// CHECK FOR VALID COLUMNS IN customCols - ALSO PREVENTS INJECTIONS IN KEYS
	if(customCols != nil){
		for key, val := range customCols {
			if info, ok := customAccountInfo[key]; !ok {
				return errors.New("Incorrect data supplied!");
			}else{
				_, err := convertDataToString(dataTypes[info.dataType], val);
				if(err != nil){ return err; }
			}
		}
	}

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(interface{}), new(interface{})};

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select "+usersColumnID+", "+usersColumnPassword+", ";
	if(customCols != nil){
		for key, _ := range customCols {
			selectQuery = selectQuery+key+", ";
			//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
			vals = append(vals, new(interface{}));
		}
	}
	selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+usersColumnName+"=\""+userName+"\";";

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
	//dbIndex := *(vals[0]).(*int); // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*interface{});

	//COMPARE HASHED PASSWORDS
	if(!helpers.CheckPasswordHash(password, dbPass.([]byte))){
		return errors.New("Login or password is incorrect");
	}

	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CUSTOM LOGIN CLIENT   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to log in a
// client when using the SQL features.
func CustomLoginClient(password string, customCols map[string]interface{}) error {
	if(len(password) == 0){
		return errors.New("A password is required to log in");
	}

	// CHECK FOR VALID COLUMNS IN customCols - ALSO PREVENTS INJECTIONS IN KEYS
	if(customCols != nil){
		if(len(customCols) == 0){
			return errors.New("Custom data is required to log in");
		}
		for key, val := range customCols {
			if info, ok := customAccountInfo[key]; !ok {
				return errors.New("Incorrect data supplied!");
			}else{
				_, err := convertDataToString(dataTypes[info.dataType], val);
				if(err != nil){ return err; }
			}
		}
	}else{
		return errors.New("Custom data is required to log in");
	}

	//FIRST THREE ARE id, userName, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(interface{}), new(interface{}), new(interface{})};
	var customKeys []string = []string{};

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select "+usersColumnID+", "+usersColumnName+", "+usersColumnPassword+", ";
	for key, _ := range customCols {
		selectQuery = selectQuery+key+", ";
		//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
		vals = append(vals, new(interface{}));
		customKeys = append(customKeys, key);
	}
	selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+customKeys[0]+"=";
	if(dataTypeNeedsQuotes(customAccountInfo[customKeys[0]].dataType)){
		selectQuery = selectQuery+"\""+customCols[customKeys[0]].(string)+"\";";
	}else{
		selectQuery = selectQuery+customCols[customKeys[0]].(string)+";";
	}


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
	//dbIndex := *(vals[0]).(*int); // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[2]).(*interface{});

	//COMPARE HASHED PASSWORDS
	if(!helpers.CheckPasswordHash(password, dbPass.([]byte))){
		return errors.New("Login or password is incorrect");
	}

	fmt.Println("logging in as:", string((*(vals[1]).(*interface{})).([]byte)));

	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHANGE PASSWORD   ///////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to log in a
// client when using the SQL features.
func ChangePassword(userName string, password string, customCols map[string]interface{}) error {
	if(len(userName) == 0){
		return errors.New("A user name is required to change a password");
	}else if(len(password) == 0){
		return errors.New("A password is required to change a password");
	}else if(checkStringSQLInjection(userName)){
		return errors.New("Malicious characters detected");
	}

	// CHECK FOR VALID COLUMNS IN customCols - ALSO PREVENTS INJECTIONS IN KEYS
	if(customCols != nil){
		for key, _ := range customCols {
			if _, ok := customAccountInfo[key]; !ok {
				return errors.New("Incorrect data supplied!");
			}
		}
	}

	//PASSWORD VERIFICATION & DATABASE RETRIEVAL

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(interface{}), new(interface{})};
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
	selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+usersColumnName+"=\""+userName+"\";";

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery);
	if(err != nil){ return err; }
	//
	checkRows.Next();
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close();
		return errors.New("User name or password is incorrect");
	}
	checkRows.Close();

	//
	dbIndex := *(vals[0]).(*interface{}); // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*interface{});

	//COMPARE HASHED PASSWORDS
	if(!helpers.CheckPasswordHash(password, dbPass.([]byte))){
		return errors.New("User name or password is incorrect");
	}

	//UPDATE THE PASSWORD
	_, updateErr := database.Exec("UPDATE "+tableUsers+" SET "+usersColumnPassword+" WHERE "+usersColumnID+"="+strconv.Itoa(dbIndex.(int))+";");
	if(updateErr != nil){ return updateErr; }

	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   CHANGE ACCOUNT INFO   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. Use the client APIs to log in a
// client when using the SQL features.
func ChangeAccountInfo(userName string, password string, customCols map[string]interface{}) error {
	if(len(userName) == 0){
		return errors.New("A user name is required to change account info");
	}else if(len(password) == 0){
		return errors.New("A password is required to change account info");
	}else if(checkStringSQLInjection(userName)){
		return errors.New("Malicious characters detected");
	}

	// CHECK FOR VALID COLUMNS IN customCols - ALSO PREVENTS INJECTIONS IN KEYS
	if(customCols != nil){
		if(len(customCols) == 0){
			return errors.New("New account info data is required to change account info");
		}
		for key, val := range customCols {
			if info, ok := customAccountInfo[key]; !ok {
				return errors.New("Incorrect data supplied!");
			}else{
				_, err := convertDataToString(dataTypes[info.dataType], val);
				if(err != nil){ return err; }
			}
		}
	}else{
		return errors.New("New account info data is required to change account info");
	}

	//PASSWORD VERIFICATION & DATABASE RETRIEVAL

	//FIRST TWO ARE id, password IN THAT ORDER
	var vals []interface{} = []interface{}{new(interface{}), new(interface{})};
	var valsList []interface{} = []interface{}{};

	//CONSTRUCT SELECT QUERY
	selectQuery := "Select "+usersColumnID+", "+usersColumnPassword+", ";
	for key, val := range customCols {
		selectQuery = selectQuery+key+", ";
		//MAINTAIN THE ORDER IN WHICH THE COLUMNS WERE DECLARED VIA A SLICE
		vals = append(vals, new(interface{}));
		valsList = append(valsList, []interface{}{val, customAccountInfo[key].dataType, key});
	}
	selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+usersColumnName+"=\""+userName+"\";";

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery);
	if(err != nil){ return err; }
	//
	checkRows.Next();
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close();
		return errors.New("User name or password is incorrect");
	}
	checkRows.Close();

	//
	dbIndex := *(vals[0]).(*interface{}); // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*interface{});

	//COMPARE HASHED PASSWORDS
	if(!helpers.CheckPasswordHash(password, dbPass.([]byte))){
		return errors.New("User name or password is incorrect");
	}

	//UPDATE THE AccountInfoColumns

	//MAKE UPDATE QUERY
	updateQuery := "UPDATE "+tableUsers+" SET ";
	for i := 0; i < len(valsList); i++ {
		dt := valsList[i].([]interface{})[1].(int);
		//GET STRING VALUE & CHECK FOR INJECTIONS
		value, valueErr := convertDataToString(dataTypes[dt], valsList[i].([]interface{})[0]);
		if(valueErr != nil){ return valueErr; }
		//CHECK IF DATA TYPE NEEDS QUOTES
		if(dataTypeNeedsQuotes(dt)){
			updateQuery = updateQuery+valsList[i].([]interface{})[2].(string)+"=\""+value+"\", ";
		}else{
			updateQuery = updateQuery+valsList[i].([]interface{})[2].(string)+"="+value+", ";
		}
	}
	updateQuery = updateQuery[0:len(updateQuery)-2]+" WHERE "+usersColumnID+"="+strconv.Itoa(dbIndex.(int))+";";

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
// client's account when using the SQL features.
func DeleteAccount(userName string, password string, customCols map[string]interface{}) error {
	if(len(userName) == 0){
		return errors.New("A user name is required to delete an account");
	}else if(len(password) == 0){
		return errors.New("A password is required to delete an account");
	}else if(checkStringSQLInjection(userName)){
		return errors.New("Malicious characters detected");
	}

	// CHECK FOR VALID COLUMNS IN customCols - ALSO PREVENTS INJECTIONS IN KEYS
	if(customCols != nil){
		for key, val := range customCols {
			if info, ok := customAccountInfo[key]; !ok {
				return errors.New("Incorrect data supplied!");
			}else{
				_, err := convertDataToString(dataTypes[info.dataType], val);
				if(err != nil){ return err; }
			}
		}
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
	selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+usersColumnName+"=\""+userName+"\";";

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery);
	if(err != nil){ return err; }
	//
	checkRows.Next();
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		checkRows.Close();
		return errors.New("User name or password is incorrect");
	}
	checkRows.Close();

	//
	dbIndex := *(vals[0]).(*int); // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*[]byte);

	//COMPARE HASHED PASSWORDS
	if(!helpers.CheckPasswordHash(password, dbPass)){
		return errors.New("User name or password is incorrect");
	}

	//EVERYTHING WENT FINE AND DANDY, DELETE THE ACCOUNT
	_, deleteErr := database.Exec("DELETE FROM "+tableUsers+" WHERE "+usersColumnID+"="+strconv.Itoa(dbIndex)+";");
	if(deleteErr != nil){ return deleteErr; }

	//REMOVE INSTANCES FROM friends TABLE
	database.Exec("DELETE FROM "+tableFriends+" WHERE "+friendsColumnUser+"="+strconv.Itoa(dbIndex)+" OR "+friendsColumnFriend+"="+strconv.Itoa(dbIndex)+";");

	//
	return nil;
}
