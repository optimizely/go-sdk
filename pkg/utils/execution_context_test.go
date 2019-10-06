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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCancelableExecutionCtx(t *testing.T) {

	var exeCtx ExecutionCtx
	exeCtx = NewCancelableExecutionCtx()

	assert.True(t, reflect.TypeOf(exeCtx) == reflect.TypeOf(&CancelableExecutionCtx{}))
	assert.NotNil(t, exeCtx.GetContext())
	assert.NotNil(t, exeCtx.GetWaitSync())
}

func TestTerminateAndWait(t *testing.T) {

	exeCtx := &CancelableExecutionCtx{}
	exeCtx.TerminateAndWait()
	assert.Nil(t, exeCtx.CancelFunc)

	exeCtx = NewCancelableExecutionCtx()
	exeCtx.TerminateAndWait()
	assert.NotNil(t, exeCtx.CancelFunc)
}
