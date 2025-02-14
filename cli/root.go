// Copyright © 2017 RooFoods LTD
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"os"

	"github.com/deliveroo/paddle/cli/data"
	"github.com/deliveroo/paddle/cli/pipeline"
	"github.com/deliveroo/paddle/cli/steps"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var outputVersion bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "paddle",
	Short: "Canoe tool for data archival and processing",
	Long:  "Canoe tool for data archival and processing",
	Run: func(cmd *cobra.Command, args []string) {
		if outputVersion {
			fmt.Println(PaddleVersion)
		} else {
			cmd.Help()
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.paddle.yaml)")
	RootCmd.Flags().BoolVar(&outputVersion, "version", false, "output paddle version")

	RootCmd.AddCommand(data.DataCmd)
	RootCmd.AddCommand(pipeline.PipelineCmd)
	RootCmd.AddCommand(steps.StepsCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".paddle" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".paddle")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
