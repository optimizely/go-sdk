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

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/spf13/cobra"
)

var getEnabledFeaturesCmd = &cobra.Command{
	Use:   "get_enabled_features",
	Short: "Get enabled features",
	Long:  `Returns enabled features`,
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

		features, _ := client.GetEnabledFeatures(user)
		fmt.Printf("Enabled features for \"%s\": %v\n", userID, features)
	},
}

func init() {
	rootCmd.AddCommand(getEnabledFeaturesCmd)
	getEnabledFeaturesCmd.Flags().StringVarP(&userID, "userId", "u", "", "user id")
	getEnabledFeaturesCmd.MarkFlagRequired("userId")
}
