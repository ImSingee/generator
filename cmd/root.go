/*
Copyright © 2020 Singee <i@singee.me>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"text/template"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "generate",
	Short: "A code generator for go",
	Long: `Please use 'go:generate' to run this app. 
	
	Add this to your source file (struct_name.go, for example):
		// go:generate generate getter -t StructName

	And run this:
		go generate

	Then a file "struct_name_getter.go" will be generated
	`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		_ = os.Chdir(viper.GetString("workdir"))

		if viper.GetBool("debug") {
			info, err := GetBasicInfo()

			if err != nil {
				return err
			} else {
				fmt.Println(info)
			}
		}

		return nil
	},
}

func GetBasicInfo() (string, error) {
	const outputTemplate = `Environments:
    $GOARCH:  {{ .GOARCH }}
        The execution architecture (arm, amd64, etc.)
    $GOOS:    {{ .GOOS }}
        The execution operating system (linux, windows, etc.)
    $GOFILE:  {{ .GOFILE }}
        The base name of the file.
    $GOLINE:  {{ .GOLINE }}
        The line number of the directive in the source file.
    $GOPACKAGE:  {{ .GOPACKAGE }}
        The name of the package of the file containing the directive.

Working Directory:
    {{ .WORKDIR }}

Command Args:
    {{- range $i, $arg := .ARGS }}
    [{{ $i }}] {{ $arg -}}
    {{ end }}

Configs:
    {{- range $key, $value := .VIPER }}
    [{{ $key }}]: {{ $value -}}
    {{ end }}
`

	t, err := template.New("").Parse(outputTemplate)

	if err != nil {
		return "", err
	}

	buf := bytes.NewBufferString("")

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	err = t.Execute(buf, map[string]interface{}{
		"GOARCH":    os.Getenv("GOARCH"),
		"GOOS":      os.Getenv("GOOS"),
		"GOFILE":    os.Getenv("GOFILE"),
		"GOLINE":    os.Getenv("GOLINE"),
		"GOPACKAGE": os.Getenv("GOPACKAGE"),
		"WORKDIR":   wd,
		"ARGS":      os.Args,
		"VIPER":     viper.AllSettings(),
	})

	s := buf.String()

	return s, nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("gofile", "", "", "mock Environment value")
	rootCmd.PersistentFlags().StringP("gopackage", "", "", "mock Environment value")
	rootCmd.PersistentFlags().StringP("workdir", "w", ".", "work directory")

	rootCmd.PersistentFlags().BoolP("debug", "", false, "debug mode")

	_ = viper.BindPFlags(rootCmd.PersistentFlags())
	_ = viper.BindEnv("GOARCH")
	_ = viper.BindEnv("GOOS")
	_ = viper.BindEnv("GOFILE")
	_ = viper.BindEnv("GOLINE")
	_ = viper.BindEnv("GOPACKAGE")
}
