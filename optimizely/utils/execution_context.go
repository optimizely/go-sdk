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
	"sync"

	"github.com/optimizely/go-sdk/optimizely/logging"
)

var exeCtx = logging.GetLogger("ExecutionCtx")

// ExecutionCtx is the interface, user can overwrite it
type ExecutionCtx interface {
	TerminateAndWait()
	GetContext() context.Context
	GetWaitSync() *sync.WaitGroup
}

// CancelableExecutionCtx has WithCancel implementation
type CancelableExecutionCtx struct {
	Wg         *sync.WaitGroup
	Ctx        context.Context
	CancelFunc context.CancelFunc
}

// NewCancelableExecutionCtxExecutionCtx returns constructed object
func NewCancelableExecutionCtxExecutionCtx() *CancelableExecutionCtx {
	ctx, cancelFn := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	return &CancelableExecutionCtx{Wg: &wg, Ctx: ctx, CancelFunc: cancelFn}
}

// TerminateAndWait sends termination signal and waits
func (ctx CancelableExecutionCtx) TerminateAndWait() {

	if ctx.CancelFunc == nil {
		exeCtx.Error("failed to shut down Execution Context properly", nil)
		return
	}
	ctx.CancelFunc()
	ctx.Wg.Wait()
}

// GetContext is context getter
func (ctx CancelableExecutionCtx) GetContext() context.Context {
	return ctx.Ctx
}

// GetWaitSync is waitgroup getter
func (ctx CancelableExecutionCtx) GetWaitSync() *sync.WaitGroup {
	return ctx.Wg
}
