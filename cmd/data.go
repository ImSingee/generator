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
	"fmt"
	"github.com/ImSingee/god/generator"
	"github.com/ImSingee/god/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getterCmd represents the getter command
var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Generate getter and setter functions for specific struct",
	RunE:  runData,
}

func init() {
	rootCmd.AddCommand(dataCmd)

	dataCmd.Flags().StringSliceP("struct", "t", []string{}, "Name list for structs")

	_ = viper.BindPFlags(dataCmd.Flags())
}

func runData(cmd *cobra.Command, args []string) error {
	structs, err := utils.GetStructsFromPackage()

	if err != nil {
		return err
	}

	t := utils.GetTemplate("filename", viper.GetString("filename"))

	// Generate Getters

	results, err := generator.GenerateGetters(structs)

	if err != nil {
		return err
	}

	for s, result := range results {
		filename := utils.ExecuteTemplate(t, map[string]interface{}{
			"struct": s,
			"type":   "getter",
		})

		err := utils.SaveGoCodeToFile(filename, result)

		if err != nil {
			return fmt.Errorf("cannot save to file %s: %w", filename, err)
		}

		fmt.Printf("Generate getter for struct %s, save as %s\n", s.Name, filename)
	}

	// Generate Setters

	results, err = generator.GenerateSetters(structs)

	if err != nil {
		return err
	}

	for s, result := range results {
		filename := utils.ExecuteTemplate(t, map[string]interface{}{
			"struct": s,
			"type":   "setter",
		})

		err := utils.SaveGoCodeToFile(filename, result)

		if err != nil {
			return fmt.Errorf("cannot save to file %s: %w", filename, err)
		}

		fmt.Printf("Generate setter for struct %s, save as %s\n", s.Name, filename)
	}

	return nil
}
