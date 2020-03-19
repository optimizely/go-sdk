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

// Package utils //
package utils

import (
	"context"
	"github.com/optimizely/go-sdk/pkg/logging"
	"sync"
	"testing"
)

func TestWithContextCancelFunc(t *testing.T) {

	ctx, cancelFunc := context.WithCancel(context.Background())
	eg := NewExecGroup(logging.GetLogger("", ""), ctx)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	eg.Go(func(ctx context.Context) {
		<-ctx.Done()
		wg.Done()
	})

	cancelFunc()
	wg.Wait()
}

func TestTerminateAndWait(t *testing.T) {

	eg := NewExecGroup(logging.GetLogger("", ""), context.Background())

	wg := &sync.WaitGroup{}
	wg.Add(1)
	eg.Go(func(ctx context.Context) {
		<-ctx.Done()
		wg.Done()
	})

	eg.TerminateAndWait()
	wg.Wait()
}
