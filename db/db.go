package db

import (
	"database/sql"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	var err error

	connString := os.Getenv("POSTGRES_DSN")
	if connString == "" {
		panic("`POSTGRES_DSN` env not set")
	}

	// config sql connection
	sqlConn, err := sql.Open("pgx", connString)
	if err != nil {
		panic(err)
	}
	sqlConn.SetMaxIdleConns(10)
	sqlConn.SetMaxIdleConns(100)
	sqlConn.SetConnMaxLifetime(time.Second * 30)
	sqlConn.SetConnMaxIdleTime(time.Second * 15)

	// initialize gorm
	postgresConn := postgres.New(postgres.Config{Conn: sqlConn})
	db, err = gorm.Open(postgresConn, &gorm.Config{})
	if err != nil {
		panic(err)
	}
}

func New() *gorm.DB {
	return db
}
