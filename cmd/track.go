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

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "track event",
	Long:  `Tracks a conversion event with eventKey`,
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

		err = client.Track(eventKey, user, map[string]interface{}{})
		if err == nil {
			fmt.Printf("Tracked event \"%s\" for \"%s\"", eventKey, userID)
		} else {
			fmt.Printf("Failed to Track event \"%s\" for \"%s\"", eventKey, userID)
		}
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
	trackCmd.Flags().StringVarP(&userID, "userId", "u", "", "user id")
	trackCmd.MarkFlagRequired("userId")
	trackCmd.Flags().StringVarP(&eventKey, "eventKey", "e", "", "event key to track")
	trackCmd.MarkFlagRequired("eventKey")
}
