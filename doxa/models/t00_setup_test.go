// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"fmt"
	"os"
	"testing"

	"github.com/doxa-erp/doxa/doxa/tools/logging"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var dbArgs = struct {
	Driver   string
	User     string
	Password string
	DB       string
	Debug    string
}{}

var testAdapter dbAdapter

func TestMain(m *testing.M) {
	initializeTests()
	res := m.Run()
	tearDownTests()
	os.Exit(res)
}

func initializeTests() {
	fmt.Printf("Initializing database for models\n")
	dbArgs.Driver = os.Getenv("DOXA_DB_DRIVER")
	if dbArgs.Driver == "" {
		dbArgs.Driver = "postgres"
	}
	dbArgs.User = os.Getenv("DOXA_DB_USER")
	if dbArgs.User == "" {
		dbArgs.User = "doxa"
	}
	dbArgs.Password = os.Getenv("DOXA_DB_PASSWORD")
	if dbArgs.Password == "" {
		dbArgs.Password = "doxa"
	}
	prefix := os.Getenv("DOXA_DB_PREFIX")
	if prefix == "" {
		prefix = "doxa"
	}

	dbArgs.DB = fmt.Sprintf("%s_models_tests", prefix)
	dbArgs.Debug = os.Getenv("DOXA_DEBUG")

	viper.Set("LogLevel", "crit")
	if dbArgs.Debug != "" {
		viper.Set("LogLevel", "debug")
		viper.Set("LogStdout", true)
	}
	logging.Initialize()

	admDB := sqlx.MustConnect(dbArgs.Driver, fmt.Sprintf("dbname=postgres sslmode=disable user=%s password=%s", dbArgs.User, dbArgs.Password))
	admDB.MustExec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbArgs.DB))
	admDB.MustExec(fmt.Sprintf("CREATE DATABASE %s", dbArgs.DB))
	admDB.Close()

	DBConnect(dbArgs.Driver, ConnectionParams{
		DBName:   dbArgs.DB,
		User:     dbArgs.User,
		Password: dbArgs.Password,
		SSLMode:  "disable",
	})
	testAdapter = adapters[db.DriverName()]
}

func tearDownTests() {
	DBClose()
	fmt.Printf("Tearing down database for models\n")
	admDB := sqlx.MustConnect(dbArgs.Driver, fmt.Sprintf("dbname=postgres sslmode=disable user=%s password=%s", dbArgs.User, dbArgs.Password))
	admDB.MustExec(fmt.Sprintf("DROP DATABASE %s", dbArgs.DB))
	admDB.Close()
}
