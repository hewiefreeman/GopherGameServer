package database

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// AccountInfoColumn is the representation of an extra column on the users table that you can define. You can define as many
// as you want. These work with the ServerCallbacks and client APIs to provide you with information on data retrieved from
// the database when the corresponding callback is triggered.
//
// You can make an AccountInfoColumn unique, which means when someone tries to update or insert into a unique column, the server
// will first check if any other row has that same value in that unique column. If a unique column cannot be updated because another
// row has the same value, an error will be sent back to the client. Keep in mind, this is an expensive task and should be used lightly,
// mainly for extra authentication.
type AccountInfoColumn struct {
	dataType  int
	maxSize   int
	precision int
	notNull   bool
	unique    bool
	encrypt   bool
}

var (
	customAccountInfo map[string]AccountInfoColumn = make(map[string]AccountInfoColumn)
)

// MySQL database data types. Use one of these when making a new AccountInfoColumn or
// CustomTable's columns. The parentheses next to a type indicate it requires a maximum
// size when making a column of that type. Two pairs of parentheses means it requires a
// decimal precision number as well a max size.
const (
	//NUMERIC TYPES
	DataTypeTinyInt   = iota // TINYINT()
	DataTypeSmallInt         // SMALLINT()
	DataTypeMediumInt        // MEDIUMINT()
	DataTypeInt              // INTEGER()
	DataTypeFloat            // FLOAT()()
	DataTypeDouble           // DOUBLE()()
	DataTypeDecimal          // DECIMAL()()
	DataTypeBigInt           // BIGINT()

	//CHARACTER TYPES
	DataTypeChar            // CHAR()
	DataTypeVarChar         // VARCHAR()
	DataTypeNationalVarChar // NVARCHAR()
	DataTypeJSON            // JSON

	//TEXT TYPES
	DataTypeTinyText   // TINYTEXT
	DataTypeMediumText // MEDIUMTEXT
	DataTypeText       // TEXT()
	DataTypeLongText   // LONGTEXT

	//DATE TYPES
	DataTypeDate      // DATE
	DataTypeDateTime  // DATETIME()
	DataTypeTime      // TIME()
	DataTypeTimeStamp // TIMESTAMP()
	DataTypeYear      // YEAR()

	//BINARY TYPES
	DataTypeTinyBlob   // TINYBLOB
	DataTypeMediumBlob // MEDIUMBLOB
	DataTypeBlob       // BLOB()
	DataTypeLongBlob   // LONGBLOB
	DataTypeBinary     // BINARY()
	DataTypeVarBinary  // VARBINARY()

	//OTHER TYPES
	DataTypeBit  // BIT()
	DataTypeENUM // ENUM()
	DataTypeSet  // SET()
)

var (
	//DATA TYPES THAT REQUIRE A SIZE
	dataTypesSize []string = []string{
		"TINYINT",
		"SMALLINT",
		"MEDIUMINT",
		"INTEGER",
		"BIGINT",
		"CHAR",
		"VARCHAR",
		"NVARCHAR",
		"TEXT",
		"DATETIME",
		"TIME",
		"TIMESTAMP",
		"YEAR",
		"BLOB",
		"BINARY",
		"VARBINARY",
		"BIT",
		"ENUM",
		"SET"}

	//DATA TYPES THAT REQUIRE A SIZE AND PRECISION
	dataTypesPrecision []string = []string{
		"FLOAT",
		"DOUBLE",
		"DECIMAL"}

	//DATA TYPE LITERAL NAME LIST
	dataTypes []string = []string{
		"TINYINT",
		"SMALLINT",
		"MEDIUMINT",
		"INTEGER",
		"FLOAT",
		"DOUBLE",
		"DECIMAL",
		"BIG INT",
		"CHAR",
		"VARCHAR",
		"NVARCHAR",
		"JSON",
		"TINYTEXT",
		"MEDIUMTEXT",
		"TEXT",
		"LONGTEXT",
		"DATE",
		"DATETIME",
		"TIME",
		"TIMESTAMP",
		"YEAR",
		"TINYBLOB",
		"MEDIUMBLOB",
		"BLOB",
		"LONGBLOB",
		"BINARY",
		"VARBINARY",
		"BIT",
		"ENUM",
		"SET"}
)

// NewAccountInfoColumn makes a new AccountInfoColumn. You can only make new AccountInfoColumns before starting the server.
func NewAccountInfoColumn(name string, dataType int, maxSize int, precision int, notNull bool, unique bool, encrypt bool) error {
	if serverStarted {
		return errors.New("You can't make a new AccountInfoColumn after the server has started")
	} else if len(name) == 0 {
		return errors.New("database.NewAccountInfoColumn() requires a name")
	} else if dataType < 0 || dataType > len(dataTypes)-1 {
		return errors.New("Incorrect data type")
	} else if checkStringSQLInjection(name) {
		return errors.New("Malicious characters detected")
	}

	if isSizeDataType(dataType) && maxSize == 0 {
		return errors.New("The data type '" + dataTypesSize[dataType] + "' requires a max size")
	} else if isPrecisionDataType(dataType) && (maxSize == 0 || precision == 0) {
		return errors.New("The data type '" + dataTypesSize[dataType] + "' requires a max size and precision")
	}

	customAccountInfo[name] = AccountInfoColumn{dataType: dataType, maxSize: maxSize, precision: precision, notNull: notNull, unique: unique, encrypt: encrypt}

	//
	return nil
}

//CHECKS IF THE DATA TYPE REQUIRES A MAX SIZE
func isSizeDataType(dataType int) bool {
	for i := 0; i < len(dataTypesSize); i++ {
		if dataTypes[dataType] == dataTypesSize[i] {
			return true
		}
	}
	return false
}

//CHECKS IF THE DATA TYPE REQUIRES A MAX SIZE
func isPrecisionDataType(dataType int) bool {
	for i := 0; i < len(dataTypesPrecision); i++ {
		if dataTypes[dataType] == dataTypesPrecision[i] {
			return true
		}
	}
	return false
}

//CONVERTS DATA TYPES TO STRING FOR SQL QUERIES
func convertDataToString(dataType string, data interface{}) (string, error) {
	switch data.(type) {
	case int:
		if dataType != "INTEGER" && dataType != "TINYINT" && dataType != "MEDIUMINT" && dataType != "BIGINT" && dataType != "SMALLINT" {
			return "", errors.New("Mismatched data types")
		}
		return strconv.Itoa(data.(int)), nil

	case float32:
		if dataType != "REAL" && dataType != "FLOAT" && dataType != "DOUBLE" && dataType != "DECIMAL" {
			return "", errors.New("Mismatched data types")
		}
		return fmt.Sprintf("%f", data.(float32)), nil

	case float64:
		if dataType != "REAL" && dataType != "FLOAT" && dataType != "DOUBLE" && dataType != "DECIMAL" {
			return "", errors.New("Mismatched data types")
		}
		return strconv.FormatFloat(data.(float64), 'f', -1, 64), nil

	case string:
		if dataType != "CHAR" && dataType != "VARCHAR" && dataType != "NVARCHAR" && dataType != "JSON" && dataType != "TEXT" &&
			dataType != "TINYTEXT" && dataType != "MEDIUMTEXT" && dataType != "LONGTEXT" && dataType != "DATE" &&
			dataType != "DATETIME" && dataType != "TIME" && dataType != "TIMESTAMP" && dataType != "YEAR" {
			return "", errors.New("Mismatched data types")
		} else if checkStringSQLInjection(data.(string)) {
			return "", errors.New("Malicious characters detected")
		}
		return "\"" + data.(string) + "\"", nil

	default:
		return "", errors.New("Data type is not supported. You can open an issue on GitHub to request support for an unsupported SQL data type.")
	}
}

//CHECKS IF THERE ARE ANY MALICIOUS CHARACTERS IN A STRING
func checkStringSQLInjection(inputStr string) bool {
	return (strings.Contains(inputStr, "\"") || strings.Contains(inputStr, ")") || strings.Contains(inputStr, "(") || strings.Contains(inputStr, ";"))
}
