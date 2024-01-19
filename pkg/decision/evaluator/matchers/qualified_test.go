/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package matchers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

type QualifiedTestSuite struct {
	suite.Suite
	mockLogger *MockLogger
	matcher    Matcher
}

func (s *QualifiedTestSuite) SetupTest() {
	s.mockLogger = new(MockLogger)
	s.matcher, _ = Get(QualifiedMatchType)
}

func (s *QualifiedTestSuite) TestQualifiedMatcherNonString() {
	user := entities.UserContext{
		QualifiedSegments: []string{"a", "b", "c"},
	}

	condition := entities.Condition{
		Value: 42,
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.UnsupportedConditionValue.String(), condition.StringRepresentation))
	result, err := s.matcher(condition, user, s.mockLogger)
	s.Error(err)
	s.False(result)

	condition = entities.Condition{
		Value: false,
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.UnsupportedConditionValue.String(), condition.StringRepresentation))
	result, err = s.matcher(condition, user, s.mockLogger)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())

	condition = entities.Condition{
		Value: []string{},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.UnsupportedConditionValue.String(), condition.StringRepresentation))
	result, err = s.matcher(condition, user, s.mockLogger)
	s.Error(err)
	s.False(result)

	condition = entities.Condition{
		Value: map[string]interface{}{},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.UnsupportedConditionValue.String(), condition.StringRepresentation))
	result, err = s.matcher(condition, user, s.mockLogger)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *QualifiedTestSuite) TestQualifiedMatcherIncorrect() {
	user := entities.UserContext{
		QualifiedSegments: []string{"a", "b", "c"},
	}

	condition := entities.Condition{
		Value: "d",
	}
	result, err := s.matcher(condition, user, s.mockLogger)
	s.NoError(err)
	s.False(result)
	s.mockLogger.AssertNotCalled(s.T(), "Warning", fmt.Sprintf(logging.UnsupportedConditionValue.String(), condition.StringRepresentation))
}

func (s *QualifiedTestSuite) TestQualifiedMatcherCorrect() {
	user := entities.UserContext{
		QualifiedSegments: []string{"a", "b", "c"},
	}

	condition := entities.Condition{
		Value: "a",
	}
	result, err := s.matcher(condition, user, s.mockLogger)
	s.NoError(err)
	s.True(result)
	s.mockLogger.AssertNotCalled(s.T(), "Warning", fmt.Sprintf(logging.UnsupportedConditionValue.String(), condition.StringRepresentation))
}

func TestQualifiedTestSuite(t *testing.T) {
	suite.Run(t, new(QualifiedTestSuite))
}
