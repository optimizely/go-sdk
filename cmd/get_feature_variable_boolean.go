/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package cmd

import (
	"fmt"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/spf13/cobra"
)

var getFeatureVariableBooleanCmd = &cobra.Command{
	Use:   "get_feature_variable_boolean",
	Short: "Get feature variable boolean",
	Long:  `Returns boolean feature variable`,
	Run: func(cmd *cobra.Command, args []string) {
		optimizelyFactory := &client.OptimizelyFactory{
			SDKKey: sdkKey,
		}

		client, err := optimizelyFactory.StaticClient()

		if err != nil {
			fmt.Printf("Error instantiating client: %s\n", err)
			return
		}

		user := entities.UserContext{
			ID:         userID,
			Attributes: map[string]interface{}{},
		}

		value, _ := client.GetFeatureVariableBoolean(featureKey, variableKey, user)
		if err == nil {
			fmt.Printf("Feature \"%s\" variable \"%s\" boolean value: %t\n", featureKey, variableKey, value)
		} else {
			fmt.Printf("Get feature \"%s\" variable \"%s\" boolean failed with error: %s\n", featureKey, variableKey, err)
		}
	},
}

func init() {
	rootCmd.AddCommand(getFeatureVariableBooleanCmd)
	getFeatureVariableBooleanCmd.Flags().StringVarP(&userID, "userId", "u", "", "user id")
	getFeatureVariableBooleanCmd.MarkFlagRequired("userId")
	getFeatureVariableBooleanCmd.Flags().StringVarP(&featureKey, "featureKey", "f", "", "feature key for feature")
	getFeatureVariableBooleanCmd.MarkFlagRequired("featureKey")
	getFeatureVariableBooleanCmd.Flags().StringVarP(&variableKey, "variableKey", "v", "", "variable key for feature variable")
	getFeatureVariableBooleanCmd.MarkFlagRequired("variableKey")
}
