package database

import (
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"database/sql"
	"errors"
	"strconv"
	"sync"
)

var (
	shardingInit bool
	partitionSplit int = 500 // Each database shard will split into tables of this size for easy further sharding later on
	shardTargetLoad int = 30000 // The maximum load a database shard should take in your system. This does not control the connection, merely triggers a warning for admins

	// Table name prefixes
	usersPrefix string = "users_"
	friendsPrefix string = "friends_"
	autologPrefix string = "autologs_"

	// Database shards
	userShardsMux sync.Mutex
	userShards    map[int]*DBShard = make(map[int]*DBShard) // default trusted database shards

	friendsShardsMux sync.Mutex
	friendsShards    map[int]*DBShard = make(map[int]*DBShard) // dedicated friends database shards

	autologShardsMux sync.Mutex
	autologShards    map[int]*DBShard = make(map[int]*DBShard) // dedicated autolog database shards
)

type DBShard struct {
	master *DBReplica

	mux       sync.Mutex
	replicaOn int
	replicas  []*DBReplica
}

type DBReplica struct {
	conn *sql.DB

	ip       string
	port     int
	protocol string
	user     string
	password string

	mux     sync.Mutex
	healthy bool
}

/////////////////////////////////////////////////////////////////////////////////////////
//////   INIT   /////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func setShardingDefaults(ip string, port int, protocol string, userName string, password string) error {
	// Set default shard if no shards are set
	if len(userShards) == 0 {
		// Make connection to default shard
		defShard, shardErr := NewDBShard(ip, port, protocol, userName, password, false)
		if shardErr {
			return shardErr
		}
		// Append default shard to user shards
		userShardsMux.Lock()
		appendShards(&userShards, defShard)
		userShardsMux.Unlock()
	}

	//
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//////   SET-UP   ///////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func AddUserShards(shards *DBShard...) {
	// REQUIRES GLOBAL PAUSE (Which only master server can do)

	userShardsMux.Lock()
	//prevLength := len(userShards)
	appendShards(&userShards, shards)
	userShardsMux.Unlock()

	// IF THIS IS MASTER, RESORT THE DATABASES
	// THEN SEND COMMAND TO REST OF GAME SERVERS
}

func AddFriendsShards(shards *DBShard...) {
	// REQUIRES GLOBAL PAUSE (Which only master server can do)

	friendsShardsMux.Lock()
	//prevLength := len(userShards)
	appendShards(&friendsShards, shards)
	friendsShardsMux.Unlock()

	// IF THIS IS MASTER, RESORT THE DATABASES
	// THEN SEND COMMAND TO REST OF GAME SERVERS
}

func AddAutologShards(shards *DBShard...) {
	// REQUIRES GLOBAL PAUSE (Which only master server can do)

	autologShardsMux.Lock()
	//prevLength := len(userShards)
	appendShards(&autologShards, shards)
	autologShardsMux.Unlock()

	// IF THIS IS MASTER, RESORT THE DATABASES
	// THEN SEND COMMAND TO REST OF GAME SERVERS
}

func appendShards(dest *map[int]*DBShard, shards *DBShard...) {
	for s := range shards {
		(*dest)[len(*dest)] = s
	}
}

/////////////////////////////////////////////////////////////////////////////////////////
//////   Sharding Utilities   ///////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func GetUserShard(userName string) (*DBShard, error) {
	userShardsMux.Lock()
	userDBNumb := helpers.HashNumber(userName, len(userShards))
	var ok bool
	var s *DBShard
	if s, ok = userShards[userDBNumb]; !ok {
		userShardsMux.Unlock()
		return nil, errors.New("Cannot find shard")
	}
	userShardsMux.Unlock()
	return s, nil
}

func GetUserShardByNumber(hashNumber int) (*DBShard, error) {
	userShardsMux.Lock()
	var ok bool
	var s *DBShard
	if s, ok = userShards[hashNumber]; !ok {
		userShardsMux.Unlock()
		return nil, errors.New("Cannot find shard")
	}
	userShardsMux.Unlock()
	return s, nil
}

func GetFriendsShard(hashNumber int) (*DBShard, error) {
	friendsShardsMux.Lock()
	if len(friendsShards) == 0 {
		friendsShardsMux.Unlock()
		return getUserShardByNumber(hashNumber)
	}
	var ok bool
	var s *DBShard
	if s, ok = friendsShards[hashNumber]; !ok {
		friendsShardsMux.Unlock()
		return nil, errors.New("Cannot find shard")
	}
	friendsShardsMux.Unlock()
	return s, nil
}

func GetAutoLogShard(hashNumber int) (*DBShard, error) {
	autologShardsMux.Lock()
	if len(autologShards) == 0 {
		autologShardsMux.Unlock()
		return getUserShardByNumber(hashNumber)
	}
	var ok bool
	var s *DBShard
	if s, ok = autologShards[hashNumber]; !ok {
		autologShardsMux.Unlock()
		return nil, errors.New("Cannot find shard")
	}
	autologShardsMux.Unlock()
	return s, nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//////   DBShard   //////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func NewDBShard(ip string, port int, protocol string, userName string, password string, master bool) (*DBShard, error) {
	replica, replicaErr := NewDBReplica(ip, port, protocol, userName, password)
	if replicaErr != nil {
		return nil, replicaErr
	}
	if master {
		shard := DBShard{master: replica}
		return &shard, nil
	}
	replicaList := []*DBReplica{replica}
	shard := DBShard{replicas: replicaList}
	return &shard, nil
}

func (s *DBShard) SetMasterDB(r *DBReplica) {
	if inited {
		return
	}
	s.master = r
}

func (s *DBShard) AddReplica(r *DBReplica) {
	// REQUIRES GLOBAL PAUSE (Which only master server can do)

	s.mux.Lock()
	s.replicas = append(s.replicas, r)
	s.mux.Unlock()

	// IF THIS IS MASTER, SEND COMMAND TO REST OF GAME SERVERS
}

func (s *DBShard) Master() *DBReplica {
	return s.master
}

func (s *DBShard) GetReplica() *DBReplica {
	s.mux.Lock()
	// GET NEXT HEALTHY REPLICA
	for i := 0; i < len(s.replicas); i++ {
		s.replicaOn++
		if s.replicaOn > len(s.replicas)-1 {
			s.replicaOn = 0
		}
		if s.replicas[s.replicaOn].Healthy() {
			r := s.replicas[s.replicaOn]
			s.mux.Unlock()
			return r
		}
	}
	s.mux.Unlock()
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//////   DBReplica   ////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func NewDBReplica(ip string, port int, protocol string, userName string, password string) (*DBReplica, error) {
	db, err := sql.Open("mysql", userName+":"+password+"@"+protocol+"("+ip+":"+strconv.Itoa(port)+")/"+databaseName)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	r := DBReplica{conn: db, ip: ip, port: port, protocol: protocol, user:userName, password: password, healthy: true}

	return &r, nil
}

func (r *DBReplica) Conn() *sql.DB {
	return r.conn
}

func (r *DBReplica) IP() string {
	return r.ip
}

func (r *DBReplica) Port() int {
	return r.port
}

func (r *DBReplica) Protocol() string {
	return r.protocol
}

func (r *DBReplica) UserName() string {
	return r.user
}

func (r *DBReplica) Password() string {
	return r.password
}

func (r *DBReplica) Healthy() bool {
	r.mux.Lock()
	h := r.healthy
	r.mux.Unlock()
	return h
}

func (r *DBReplica) SetHealthy(h bool) {
	r.mux.Lock()
	r.healthy = h
	r.mux.Unlock()
}

/////////////////////////////////////////////////////////////////////////////////////////
//////   TABE PREFIX SETTERS   //////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func SetUsersTablePrefix(prefix string) {
	if inited {
		return
	}
	usersPrefix = prefix
}

func SetFriendsTablePrefix(prefix string) {
	if inited {
		return
	}
	friendsPrefix = prefix
}

func SetAutologTablePrefix(prefix string) {
	if inited {
		return
	}
	autologPrefix = prefix
}
