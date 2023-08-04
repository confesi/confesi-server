package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	db, err = gorm.Open(postgresConn, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // added for debugging to log the underlying Gorm-generated SQL queries
	})
	if err != nil {
		panic(err)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func New() *gorm.DB {
	return db
}
