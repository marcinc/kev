/**
 * Copyright 2020 Appvia Ltd <info@appvia.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package testutil

import (
	"bytes"
	"strings"

	"github.com/appvia/kev/pkg/kev/log"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func NewLogger(level logrus.Level) *test.Hook {
	var buffer = &bytes.Buffer{}
	log.SetOutput(buffer)
	log.SetLogLevel(level)
	return test.NewLocal(log.GetLogger())
}

func GetLoggedMsgs(hook *test.Hook) string {
	var out strings.Builder
	for _, entry := range hook.Entries {
		out.WriteString(entry.Message + "\n")
	}
	return out.String()
}

func GetLoggedLevel(hook *test.Hook) string {
	return hook.LastEntry().Level.String()
}
