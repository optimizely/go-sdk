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
)

// ExecGroup is a utility for managing graceful, blocking cancellation of goroutines.
type ExecGroup struct {
	wg         *sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc
	logger     logging.OptimizelyLogProducer
}

// NewExecGroup returns constructed object
func NewExecGroup(ctx context.Context, logger logging.OptimizelyLogProducer) *ExecGroup {
	nctx, cancelFn := context.WithCancel(ctx)
	wg := sync.WaitGroup{}

	return &ExecGroup{wg: &wg, ctx: nctx, cancelFunc: cancelFn, logger:logger}
}

// Go initiates a goroutine with the inputted function. Each invocation increments a shared WaitGroup
// before being initiated. Once the supplied function exits the WaitGroup is decremented.
// A common ctx is passed to each input function to signal a shutdown sequence.
func (c ExecGroup) Go(f func(ctx context.Context)) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		f(c.ctx)
	}()
}

// TerminateAndWait sends termination signal and waits
func (c ExecGroup) TerminateAndWait() {

	if c.cancelFunc == nil {
		c.logger.Error("failed to shut down Execution Context properly", nil)
		return
	}
	c.cancelFunc()
	c.wg.Wait()
}
