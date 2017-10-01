/**
 ****************************************************************************
 * Copyright 2017 Jonathan Barronville <jonathan@belairlabs.com>            *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *     http://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ****************************************************************************
 */

package thoruh

import (
	"reflect"
	"runtime"
	"testing"
)

func TestOptions(t *testing.T) {
	t.Run("No defined options and empty argument values.", func(t *testing.T) {
		t.Parallel()
		argumentValues := []string{}
		options := NewOptions(argumentValues)
		optionsResult := options.Parse()
		if len(argumentValues) != len(optionsResult.RemainingArgumentValues) {
			t.Fail()
		}
	})
	t.Run("No defined options and unknown options provided.", func(t *testing.T) {
		t.Parallel()
		argumentValues := []string{"-x", "--foo"}
		options := NewOptions(argumentValues)
		optionsResult := options.Parse()
		for i, parsedOption := range optionsResult.Options {
			switch i {
			case 0:
				if !parsedOption.Error {
					t.Fail()
				}
				err := newUnknownOptionParseError(options, OptionTypeShort, []rune("x"))
				if !reflect.DeepEqual(parsedOption.Value, err) {
					t.Fail()
				}
			case 1:
				if !parsedOption.Error {
					t.Fail()
				}
				err := newUnknownOptionParseError(options, OptionTypeLong, []rune("foo"))
				if !reflect.DeepEqual(parsedOption.Value, err) {
					t.Fail()
				}
			}
		}
		if len(optionsResult.RemainingArgumentValues) != 0 {
			t.Fail()
		}
		if runtime.GOOS == "windows" {
			t.Run("DOS-style slash. (windows)", func(t *testing.T) {
				t.Parallel()
				argumentValues := []string{"/x"}
				options := NewOptions(argumentValues)
				optionsResult := options.Parse()
				for i, parsedOption := range optionsResult.Options {
					if i == 0 {
						if !parsedOption.Error {
							t.Fail()
						}
						err := newUnknownOptionParseError(options, OptionTypeShort, []rune("x"))
						if !reflect.DeepEqual(parsedOption.Value, err) {
							t.Fail()
						}
					}
				}
				if len(optionsResult.RemainingArgumentValues) != 0 {
					t.Fail()
				}
			})
		}
	})
}
