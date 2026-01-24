package modules

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net"
	"sync"
)

const DBName = "app_blog"

var (
	instance *DBManager
	once     sync.Once
)

type DBManager struct {
	DB       *sql.DB
	TcpAddr  string
	TcpPort  string
	Username string
	Password string
}

func GetDBManager() *DBManager {
	once.Do(func() {
		instance = new(DBManager)
	})
	return instance
}
func (dbm *DBManager) Init() error {
	dsn, _ := dbm.dsnString()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("sql Open error:%v", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("sql Open error:%v", err)
	}
	dbm.DB = db
	return nil
}

func (dbm *DBManager) dsnString() (dsn string, err error) {
	defer func() {
		if r := recover(); r != nil {
			dsn = ""
			err = fmt.Errorf("panic in dsnString: %v", r)
		}
	}()
	if parseIp := net.ParseIP(dbm.TcpAddr); parseIp == nil {
		return "", fmt.Errorf("invalid tcp address: %s", dbm.TcpAddr)
	}
	if dbm.Username == "" {
		return "", fmt.Errorf("invalid username: %s", dbm.Username)
	}
	dsn = fmt.Sprintf("postgres://%v:%v@%v:%v/%v", dbm.Username, dbm.Password, dbm.TcpAddr, dbm.TcpPort, DBName)
	//dsn = fmt.Sprintf("%v:%v@tcp(%v:%v)/blog?charset=utf8mb4&parseTime=True", dbm.Username, dbm.Password, dbm.TcpAddr, dbm.TcpPort)
	return dsn, err
}
