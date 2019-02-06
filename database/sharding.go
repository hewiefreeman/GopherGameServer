package database

import (
	//"github.com/hewiefreeman/GopherGameServer/helpers"
	"database/sql"
	"errors"
	"strconv"
	"sync"
)

var (
	shardingInit bool
	shardTargetEntries int = 30000 // The maximum entries a database table shard should hold in your system. This does not control the max entry count
	shardPercentWarning int = 80 //

	// Table name prefixes
	usersShardPrefix string = "users_"
	friendsShardPrefix string = "friends_"
	autologShardPrefix string = "autologs_"

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

	id int

	mux       sync.Mutex
	replicaOn int
	replicas  []*DBReplica
}

type DBReplica struct {
	conn *sql.DB

	id int

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

func ShardingInited() bool {
	return shardingInit
}

func setShardingDefaults(ip string, port int, protocol string, userName string, password string) error {
	// Set default shard if no shards are set
	if len(userShards) == 0 {
		// Make connection to default partitions
		var defaultDBs []*DBShard = make([]*DBShard, 3, 3)
		for i := 0; i < 3; i++ {
			defShard, shardErr := NewDBShard(ip, port, protocol, userName, password, false, i)
			if shardErr != nil {
				return shardErr
			}
			defaultDBs[i] = defShard
		}

		// Append default shard to user shards
		userShardsMux.Lock()
		appendShards(&userShards, defaultDBs)
		userShardsMux.Unlock()
	}

	//
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//////   SET-UP   ///////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func SetShardingTargetEntries(maxEntries int) {
	if inited {
		return
	}
	shardTargetEntries = maxEntries
}

func SetEntriesWarningPercent(percent int) {
	if inited {
		return
	}
	shardPercentWarning = percent
}

func AddUserShards(shards []*DBShard) {
	// REQUIRES GLOBAL PAUSE (Which only master server can do)

	userShardsMux.Lock()
	//prevLength := len(userShards)
	appendShards(&userShards, shards)
	userShardsMux.Unlock()

	// IF THIS IS MASTER, RESORT THE DATABASES
	// THEN SEND COMMAND TO REST OF GAME SERVERS
}

func AddFriendsShards(shards []*DBShard) {
	// REQUIRES GLOBAL PAUSE (Which only master server can do)

	friendsShardsMux.Lock()
	//prevLength := len(userShards)
	appendShards(&friendsShards, shards)
	friendsShardsMux.Unlock()

	// IF THIS IS MASTER, RESORT THE DATABASES
	// THEN SEND COMMAND TO REST OF GAME SERVERS
}

func AddAutologShards(shards []*DBShard) {
	// REQUIRES GLOBAL PAUSE (Which only master server can do)

	autologShardsMux.Lock()
	//prevLength := len(userShards)
	appendShards(&autologShards, shards)
	autologShardsMux.Unlock()

	// IF THIS IS MASTER, RESORT THE DATABASES
	// THEN SEND COMMAND TO REST OF GAME SERVERS
}

func appendShards(dest *map[int]*DBShard, shards []*DBShard) {
	for i := 0; i < len(shards); i++ {
		(*shards[i]).id = len(*dest)
		(*dest)[len(*dest)] = shards[i]
	}
}

/////////////////////////////////////////////////////////////////////////////////////////
//////   Shard Getters   ////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func GetUserShard(hashNumber int) (*DBShard, error) {
	userShardsMux.Lock()
	if len(userShards) == 0 {
		userShardsMux.Unlock()
		return nil, errors.New("No shards exist")
	}
	var ok bool
	var s *DBShard
	if s, ok = userShards[hashNumber%len(userShards)]; !ok {
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
		return GetUserShard(hashNumber)
	}
	var ok bool
	var s *DBShard
	if s, ok = friendsShards[hashNumber%len(friendsShards)]; !ok {
		friendsShardsMux.Unlock()
		return nil, errors.New("Cannot find shard")
	}
	friendsShardsMux.Unlock()
	return s, nil
}

func GetAutologShard(hashNumber int) (*DBShard, error) {
	autologShardsMux.Lock()
	if len(autologShards) == 0 {
		autologShardsMux.Unlock()
		return GetUserShard(hashNumber)
	}
	var ok bool
	var s *DBShard
	if s, ok = autologShards[hashNumber%len(autologShards)]; !ok {
		autologShardsMux.Unlock()
		return nil, errors.New("Cannot find shard")
	}
	autologShardsMux.Unlock()
	return s, nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//////   DBShard   //////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func NewDBShard(ip string, port int, protocol string, userName string, password string, master bool, shardID int) (*DBShard, error) {
	replica, replicaErr := NewDBReplica(ip, port, protocol, userName, password, 0)
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
	(*r).id = len(s.replicas)
	s.replicas = append(s.replicas, r)
	s.mux.Unlock()

	// IF THIS IS MASTER, SEND COMMAND TO REST OF GAME SERVERS
}

func (s *DBShard) Master() *DBReplica {
	if s.master == nil {
		return s.GetReplica()
	}
	return s.master
}

func (s *DBShard) GetReplica() *DBReplica {
	s.mux.Lock()
	if len(s.replicas) == 0 {
		if s.master == nil {
			s.mux.Unlock()
			return nil
		}
		return s.master
	}
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

func (s *DBShard) ID() int {
	return s.id
}

/////////////////////////////////////////////////////////////////////////////////////////
//////   DBReplica   ////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func NewDBReplica(ip string, port int, protocol string, userName string, password string, replicaID int) (*DBReplica, error) {
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

func (r *DBReplica) ID() int {
	return r.id
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
//////   TABLE PREFIX SETTERS   /////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func SetUsersTablePrefix(prefix string) {
	if inited {
		return
	}
	usersShardPrefix = prefix
}

func SetFriendsTablePrefix(prefix string) {
	if inited {
		return
	}
	friendsShardPrefix = prefix
}

func SetAutologTablePrefix(prefix string) {
	if inited {
		return
	}
	autologShardPrefix = prefix
}
