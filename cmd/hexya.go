// Copyright 2017 NDP Systèmes. All Rights Reserved.
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

package cmd

import (
	"fmt"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/doxa-erp/doxa/doxa/tools/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var log *logging.Logger

// DoxaCmd is the base 'doxa' command of the commander
var DoxaCmd = &cobra.Command{
	Use:   "doxa",
	Short: "Doxa is an open source modular ERP",
	Long: `Doxa is an open source modular ERP written in Go.
It is designed for high demand business data processing while being easily customizable`,
}

func init() {
	log = logging.GetLogger("init")
	cobra.OnInitialize(initConfig)

	DoxaCmd.PersistentFlags().StringP("config", "c", "", "Alternate configuration file to read. Defaults to $HOME/.doxa/")
	viper.BindPFlag("ConfigFileName", DoxaCmd.PersistentFlags().Lookup("config"))

	DoxaCmd.PersistentFlags().StringSliceP("modules", "m", []string{"github.com/doxa-erp/doxa-base/web"}, "List of module paths to load. Defaults to ['github.com/doxa-erp/doxa-base/web']")
	viper.BindPFlag("Modules", DoxaCmd.PersistentFlags().Lookup("modules"))

	DoxaCmd.PersistentFlags().StringP("log-level", "L", "info", "Log level. Should be one of 'debug', 'info', 'warn', 'error' or 'crit'")
	viper.BindPFlag("LogLevel", DoxaCmd.PersistentFlags().Lookup("log-level"))
	DoxaCmd.PersistentFlags().String("log-file", "", "File to which the log will be written")
	viper.BindPFlag("LogFile", DoxaCmd.PersistentFlags().Lookup("log-file"))
	DoxaCmd.PersistentFlags().BoolP("log-stdout", "o", false, "Enable stdout logging. Use for development or debugging.")
	viper.BindPFlag("LogStdout", DoxaCmd.PersistentFlags().Lookup("log-stdout"))
	DoxaCmd.PersistentFlags().Bool("debug", false, "Enable server debug mode for development")
	viper.BindPFlag("Debug", DoxaCmd.PersistentFlags().Lookup("debug"))
	DoxaCmd.PersistentFlags().Bool("demo", false, "Load demo data for evaluating or tests")
	viper.BindPFlag("Demo", DoxaCmd.PersistentFlags().Lookup("demo"))

	DoxaCmd.PersistentFlags().String("data-dir", "", "Path to the directory where Doxa should store its data")
	viper.BindPFlag("DataDir", DoxaCmd.PersistentFlags().Lookup("data-dir"))

	DoxaCmd.PersistentFlags().String("db-driver", "postgres", "Database driver to use")
	viper.BindPFlag("DB.Driver", DoxaCmd.PersistentFlags().Lookup("db-driver"))
	DoxaCmd.PersistentFlags().String("db-host", "/var/run/postgresql",
		"The database host to connect to. Values that start with / are for unix domain sockets directory")
	viper.BindPFlag("DB.Host", DoxaCmd.PersistentFlags().Lookup("db-host"))
	DoxaCmd.PersistentFlags().String("db-port", "5432", "Database port. Value is ignored if db-host is not set")
	viper.BindPFlag("DB.Port", DoxaCmd.PersistentFlags().Lookup("db-port"))
	DoxaCmd.PersistentFlags().String("db-user", "", "Database user. Defaults to current user")
	viper.BindPFlag("DB.User", DoxaCmd.PersistentFlags().Lookup("db-user"))
	DoxaCmd.PersistentFlags().String("db-password", "", "Database password. Leave empty when connecting through socket")
	viper.BindPFlag("DB.Password", DoxaCmd.PersistentFlags().Lookup("db-password"))
	DoxaCmd.PersistentFlags().String("db-name", "doxa", "Database name")
	viper.BindPFlag("DB.Name", DoxaCmd.PersistentFlags().Lookup("db-name"))
	DoxaCmd.PersistentFlags().String("db-ssl-mode", "prefer", "SSL mode to connect to the database. Must be one of 'disable', 'prefer' (default), 'require', 'verify-ca' and 'verify-full'")
	viper.BindPFlag("DB.SSLMode", DoxaCmd.PersistentFlags().Lookup("db-ssl-mode"))
	DoxaCmd.PersistentFlags().String("db-ssl-cert", "", "Path to client certificate file")
	viper.BindPFlag("DB.SSLCert", DoxaCmd.PersistentFlags().Lookup("db-ssl-cert"))
	DoxaCmd.PersistentFlags().String("db-ssl-key", "", "Path to client private key file")
	viper.BindPFlag("DB.SSLKey", DoxaCmd.PersistentFlags().Lookup("db-ssl-key"))
	DoxaCmd.PersistentFlags().String("db-ssl-ca", "", "Path to certificate authority certificate(s) file")
	viper.BindPFlag("DB.SSLCA", DoxaCmd.PersistentFlags().Lookup("db-ssl-ca"))
}

func initConfig() {
	viper.SetEnvPrefix("doxa")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	cfgFile := viper.GetString("ConfigFileName")
	if runtime.GOOS != "windows" {
		viper.AddConfigPath("/etc/doxa")
	}

	osUser, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("unable to retrieve current user. Error: %s", err))
	}
	defaultDoxaDir := filepath.Join(osUser.HomeDir, ".doxa")
	viper.SetDefault("DataDir", defaultDoxaDir)
	viper.AddConfigPath(defaultDoxaDir)
	viper.AddConfigPath(".")

	viper.SetConfigName("doxa")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
}
