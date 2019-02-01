package database

import (
	"database/sql"
	//"errors"
	"strconv"
	"sync"
)

var (
	shardingInit bool

	usersPrefix string = "users_"
	usersShards map[string]*DBShard

	friendsPrefix string = "friends_"
	friendsShards map[string]*DBShard

	autologPrefix string = "autologs_"
	autologShards map[string]*DBShard
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
	if usersShards == nil {
		shard, shardErr := NewDBShard(ip, port, protocol, userName, password, false)
		if shardErr != nil {
			return shardErr
		}
		usersShards = make(map[string]*DBShard)
		//
		usersShards["a-h"] = shard
		usersShards["i-q"] = shard
		usersShards["r-z"] = shard
	} else {
		fillShardGaps(&usersShards)
	}

	if friendsShards == nil {
		friendsShards = usersShards
	} else {
		fillShardGaps(&friendsShards)
	}

	if autologShards == nil {
		autologShards = usersShards
	} else {
		fillShardGaps(&autologShards)
	}

	//
	return nil
}

func fillShardGaps(shardMap *map[string]*DBShard) {

}

/////////////////////////////////////////////////////////////////////////////////////////
//////   ADDING SHARDS   ////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func SetUserShard(letters string, shard *DBShard) error {
	if !letterFormatCheck(letters) {

	} else if letterConflict(letters, usersShards) {

	}

	//
	return nil
}

func SetFriendShard(letters string, shard *DBShard) error {
	if !letterFormatCheck(letters) {

	} else if letterConflict(letters, friendsShards) {

	}

	//
	return nil
}

func SetAutologShard(letters string, shard *DBShard) error {
	if !letterFormatCheck(letters) {

	} else if letterConflict(letters, autologShards) {

	}

	//
	return nil
}

func letterFormatCheck(letters string) bool {

	// Pass
	return true
}

func letterConflict(letters string, m map[string]*DBShard) bool {

	// Pass
	return true
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
	s.mux.Lock()
	s.replicas = append(s.replicas, r)
	s.mux.Unlock()
}

func (s *DBShard) Master() *DBReplica {
	return s.master
}

func (s *DBShard) GetReplica() *DBReplica {
	s.mux.Lock()
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
