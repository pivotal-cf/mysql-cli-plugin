package ssh

import (
	_ "github.com/go-sql-driver/mysql"

	"database/sql"
	"fmt"
	"github.com/pivotal-cf/mysql-cli-plugin/service"
)

type db struct{}

func NewDB() *db {
	return &db{}
}

func (d *db) Ping(serviceInfo *service.ServiceInfo, port int) error {
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(127.0.0.1:%d)/%s?interpolateParams=true&tls=false",
		serviceInfo.Username,
		serviceInfo.Password,
		port,
		serviceInfo.DBName,
	)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}

	defer db.Close()
	return db.Ping()
}
