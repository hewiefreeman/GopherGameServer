package database

import(
	"errors"
)

// An AccountInfoColumn is the representation of an extra column on the users table that you can define. You can define as many
// as you want. The client APIs take in an optional parameter when signing up a client that you use
// to set your AccountInfoColumn(s). When a client logs in, the server callback will provide you with a
// map[string]interface{} of your AccountInfoColumn(s) where the key is the name of the column and the
// interface is the data retrieved from the column.
type AccountInfoColumn struct {
	dataType int
	maxSize int
}

var(
	customAccountInfo map[string]AccountInfoColumn = make(map[string]AccountInfoColumn)
)

// MySQL database data types. Use one of these when making a new AccountInfoColumn or
// CustomTable's columns.
const (
	//NUMERIC TYPES
	DataTypeTinyInt = iota // TINYINT()
	DataTypeSmallInt // SMALLINT()
	DataTypeReal // REAL
	DataTypeMediumInt // MEDIUMINT()
	DataTypeInt // INT()
	DataTypeFloat // FLOAT
	DataTypeDouble // DOUBLE
	DataTypeDecimal // DECIMAL
	DataTypeBigInt // BIGINT()

	//CHARACTER TYPES
	DataTypeChar // CHAR()
	DataTypeVarChar // VARCHAR()
	DataTypeNationalVarChar // NVARCHAR()
	DataTypeJSON // JSON

	//TEXT TYPES
	DataTypeTinyText // TINYTEXT
	DataTypeMediumText // MEDIUMTEXT
	DataTypeText // TEXT()
	DataTypeLongText // LONGTEXT

	//DATE TYPES
	DataTypeDate // DATE
	DataTypeDateTime // DATETIME()
	DataTypeTime // TIME()
	DataTypeTimeStamp // TIMESTAMP()
	DataTypeYear // YEAR()

	//BINARY TYPES
	DataTypeTinyBlob // TINYBLOB
	DataTypeMediumBlob // MEDIUMBLOB
	DataTypeBlob // BLOB()
	DataTypeLongBlob // LONGBLOB
	DataTypeBinary // BINARY()
	DataTypeVarBinary // VARBINARY()

	//OTHER TYPES
	DataTypeBit // BIT()
	DataTypeENUM // ENUM()
	DataTypeSet // SET()
)

var (
	//DATA TYPES THAT REQUIRE A SIZE
	dataTypesSize []string = []string{
					"TINYINT",
					"SMALLINT",
					"MEDIUMINT",
					"INT",
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

	//DATA TYPE LITERAL NAME LIST
	dataTypes []string = []string{
					"TINYINT",
					"SMALLINT",
					"REAL",
					"MEDIUMINT",
					"INT",
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

// Use this to make a new AccountInfoColumn. You can only make new AccountInfoColumns before starting the server.
func NewAccountInfoColumn(name string, dataType int, maxSize int) error {
	if(serverStarted){
		return errors.New("You can't make a new AccountInfoColumn after the server has started");
	}else if(len(name) == 0){
		return errors.New("database.NewAccountInfoColumn() requires a name");
	}else if(dataType < 0 || dataType > len(dataTypes)-1){
		return errors.New("Incorrect data type");
	}

	if(isSizeDataType(dataType) && maxSize == 0){
		return errors.New("The data type '"+dataTypesSize[dataType]+"' requires a max size");
	}

	customAccountInfo[name] = AccountInfoColumn{dataType: dataType, maxSize: maxSize};

	//
	return nil;
}

//CHECKS IF THE DATA TYPE REQUIRES A MAX SIZE
func isSizeDataType(dataType int) bool {
	for i := 0; i < len(dataTypesSize); i++ {
		if(dataTypes[dataType] == dataTypesSize[i]){
			return true;
		}
	}
	return false;
}
