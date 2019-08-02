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

var (
	userId      string
	featurekKey string
)

var isFeatureEnabledCmd = &cobra.Command{
	Use:   "is_feature_enabled",
	Short: "Is feature enabled?",
	Long:  `Determines if a feature is enabled`,
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
			ID:         userId,
			Attributes: map[string]interface{}{},
		}

		enabled, _ := client.IsFeatureEnabled(featurekKey, user)
		fmt.Printf("Is feature \"%s\" enabled for \"%s\"? %t\n", featurekKey, userId, enabled)
	},
}

func init() {
	rootCmd.AddCommand(isFeatureEnabledCmd)
	isFeatureEnabledCmd.Flags().StringVarP(&userId, "userId", "u", "", "user id")
	isFeatureEnabledCmd.MarkFlagRequired("userId")
	isFeatureEnabledCmd.Flags().StringVarP(&featurekKey, "featureKey", "f", "", "feature key to enable")
	isFeatureEnabledCmd.MarkFlagRequired("featureKey")
}
