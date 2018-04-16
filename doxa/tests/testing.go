// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/labneco/doxa/doxa/models"
	"github.com/labneco/doxa/doxa/server"
	"github.com/labneco/doxa/doxa/tools/logging"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

var driver, user, password, prefix, debug string

// RunTests initializes the database, run the tests given by m and
// tears the database down.
//
// It is meant to be used for modules testing. Initialize your module's
// tests with:
//
//     import (
//         "testing"
//         "github.com/labneco/doxa/doxa/tests"
//     )
//
//     func TestMain(m *testing.M) {
//	       tests.RunTests(m, "my_module")
//     }
func RunTests(m *testing.M, moduleName string) {
	var res int
	defer func() {
		TearDownTests(moduleName)
		if r := recover(); r != nil {
			panic(r)
		}
		os.Exit(res)
	}()
	InitializeTests(moduleName)
	res = m.Run()

}

// InitializeTests initializes a database for the tests of the given module.
// You probably want to use RunTests instead.
func InitializeTests(moduleName string) {
	fmt.Printf("Initializing database for module %s\n", moduleName)
	driver = os.Getenv("DOXA_DB_DRIVER")
	if driver == "" {
		driver = "postgres"
	}
	user = os.Getenv("DOXA_DB_USER")
	if user == "" {
		user = "doxa"
	}
	password = os.Getenv("DOXA_DB_PASSWORD")
	if password == "" {
		password = "doxa"
	}
	prefix = os.Getenv("DOXA_DB_PREFIX")
	if prefix == "" {
		prefix = "doxa"
	}
	dbName := fmt.Sprintf("%s_%s_tests", prefix, moduleName)
	debug = os.Getenv("DOXA_DEBUG")

	viper.Set("LogLevel", "crit")
	if debug != "" {
		viper.Set("LogLevel", "debug")
		viper.Set("LogStdout", true)
	}
	logging.Initialize()

	db := sqlx.MustConnect(driver, fmt.Sprintf("dbname=postgres sslmode=disable user=%s password=%s", user, password))
	db.MustExec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	db.Close()

	models.DBConnect(driver, models.ConnectionParams{
		DBName:   dbName,
		User:     user,
		Password: password,
		SSLMode:  "disable",
	})
	models.BootStrap()
	models.SyncDatabase()
	server.LoadDataRecords()
	server.LoadDemoRecords()

	server.PostInitModules()
}

// TearDownTests tears down the tests for the given module
func TearDownTests(moduleName string) {
	models.DBClose()
	fmt.Printf("Tearing down database for module %s\n", moduleName)
	dbName := fmt.Sprintf("%s_%s_tests", prefix, moduleName)
	db := sqlx.MustConnect(driver, fmt.Sprintf("dbname=postgres sslmode=disable user=%s password=%s", user, password))
	db.MustExec(fmt.Sprintf("DROP DATABASE %s", dbName))
	db.Close()
}
