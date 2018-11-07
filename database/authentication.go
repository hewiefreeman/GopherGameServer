package database

import(
	"errors"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	//"fmt"
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

	//CHECK IF USERNAME IS NOT TAKEN
	checkRows, err := database.Query("Select "+usersColumnName+" FROM "+tableUsers+" WHERE "+usersColumnName+"='"+userName+"';");
	if(err != nil){ return err; }
	//
	checkRows.Next();
	_, colsErr := checkRows.Columns();
	if(colsErr == nil){
		return errors.New("User name is taken");
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
	selectQuery = selectQuery[0:len(selectQuery)-2]+" FROM "+tableUsers+" WHERE "+usersColumnName+"='"+userName+"';";

	//EXECUTE SELECT QUERY
	checkRows, err := database.Query(selectQuery);
	if(err != nil){ return err; }
	//
	checkRows.Next();
	if scanErr := checkRows.Scan(vals...); scanErr != nil {
		return errors.New("User name or password is incorrect");
	}
	checkRows.Close();

	//
	//dbIndex := *(vals[0]).(*int); // USE FOR SERVER CALLBACK & MAKE DATABASE RESPONSE MAP
	dbPass := *(vals[1]).(*interface{});

	//COMPARE HASHED PASSWORDS
	if(!helpers.CheckPasswordHash(password, dbPass.([]byte))){
		return errors.New("User name or password is incorrect");
	}

	//
	return nil;
}
