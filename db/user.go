package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func NewUserQuery(user, passwd string, driver, dsn string) (int, int, error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// TODO driver support, need import dirver init
	db, err := sql.Open(driver, dsn)
	rName, rPasswd, rQuota, wQuota := "", "", 0, 0
	if err != nil {
		return rQuota, wQuota, err
	}
	defer db.Close()

	// sql injection
	row := db.QueryRow("select name,passwd,read_quota,write_quota from user where name =? and passwd=?;", user, passwd)
	err = row.Scan(&rName, &rPasswd, &rQuota, &wQuota)
	return rQuota, wQuota, err
}
