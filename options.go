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
	"fmt"
	"runtime"
	"strings"
	"unicode/utf8"
)

type optionArgumentType uint

const (
	// OptionArgumentTypeNone represents the argument type for an option which doesn't accept an argument.
	OptionArgumentTypeNone optionArgumentType = iota
	// OptionArgumentTypeRequired represents the argument type for an option for which an argument is required.
	OptionArgumentTypeRequired
)

type optionType uint

const (
	// OptionTypeLong represents a "long" (e.g., "--") option type.
	OptionTypeLong optionType = iota
	// OptionTypeShort represents a "short" (e.g., "-") option type.
	OptionTypeShort
)

// Option represents a command-line option.
type Option struct {
	ArgumentType optionArgumentType
	Name         string
	Type         optionType
}

// ExtraneousOptionArgumentParseError represents an error for when an argument is provided to an option which doesn't accept an argument.
type ExtraneousOptionArgumentParseError struct {
	message        string
	name           string
	OptionArgument string
	OptionName     string
	OptionType     optionType
}

// MissingOptionArgumentParseError represents an error for when an option which expects an argument doesn't receive one.
type MissingOptionArgumentParseError struct {
	message    string
	name       string
	OptionName string
	OptionType optionType
}

// UnknownOptionParseError represents an error for when an unrecognized option is found.
type UnknownOptionParseError struct {
	message    string
	name       string
	OptionName string
	OptionType optionType
}

// ParsedOptionValue represents an option which was successfully parsed.
type ParsedOptionValue struct {
	Argument     string
	ArgumentType optionArgumentType
	Name         string
	Type         optionType
}

// ParsedOption represents either an option which was successfully parsed (ParsedOptionValue) or a parsing error.
type ParsedOption struct {
	Error bool
	Value interface{}
}

// ParseResult is the structure returned by the parser, containing the results.
// ParseResult.Options is a slice of "parsed options".
// ParseResult.RemainingArgumentValues is a slice containing the remaining argument values when parsing stopped.
type ParseResult struct {
	Options                 []ParsedOption
	RemainingArgumentValues []string
}

// Options represents an options parser.
type Options struct {
	argumentValues                    [][]rune
	longOptionDescriptors             map[string]Option
	nextArgumentValueIndex            uint
	Parsed                            *ParseResult
	shortOptionDescriptors            map[string]Option
	skipArgumentsOnNextParseIteration uint
}

func newExtraneousOptionArgumentParseError(options *Options, optionType optionType, optionName []rune, optionArgument []rune) ExtraneousOptionArgumentParseError {
	var optionPrefix string
	switch optionType {
	case OptionTypeLong:
		optionPrefix = "--"
	case OptionTypeShort:
		optionPrefix = "-"
	}
	message := fmt.Sprintf("Extraneous argument \"%s\" passed to option \"%s%s\".", string(optionArgument), optionPrefix, string(optionName))
	return ExtraneousOptionArgumentParseError{
		message:        message,
		name:           "ExtraneousOptionArgumentParseError",
		OptionArgument: string(optionArgument),
		OptionName:     string(optionName),
		OptionType:     optionType,
	}
}

func (error_ ExtraneousOptionArgumentParseError) Error() string {
	return error_.message
}

func newMissingOptionArgumentParseError(options *Options, optionType optionType, optionName []rune) MissingOptionArgumentParseError {
	var optionPrefix string
	switch optionType {
	case OptionTypeLong:
		optionPrefix = "--"
	case OptionTypeShort:
		optionPrefix = "-"
	}
	message := fmt.Sprintf("Option \"%s%s\" expects an argument.", optionPrefix, string(optionName))
	return MissingOptionArgumentParseError{
		message:    message,
		name:       "MissingOptionArgumentParseError",
		OptionName: string(optionName),
		OptionType: optionType,
	}
}

func (error_ MissingOptionArgumentParseError) Error() string {
	return error_.message
}

func newUnknownOptionParseError(options *Options, optionType optionType, optionName []rune) UnknownOptionParseError {
	var optionPrefix string
	switch optionType {
	case OptionTypeLong:
		optionPrefix = "--"
	case OptionTypeShort:
		optionPrefix = "-"
	}
	message := fmt.Sprintf("Option \"%s%s\" is unknown.", optionPrefix, string(optionName))
	return UnknownOptionParseError{
		message:    message,
		name:       "UnknownOptionParseError",
		OptionName: string(optionName),
		OptionType: optionType,
	}
}

func (error_ UnknownOptionParseError) Error() string {
	return error_.message
}

// NewOptions creates a new instance of Options and returns the pointer to it.
func NewOptions(argumentValues []string) *Options {
	_argumentValues := make([][]rune, len(argumentValues))
	for argumentIndex, argumentValue := range argumentValues {
		_argumentValues[argumentIndex] = []rune(argumentValue)
	}
	options := Options{
		argumentValues:                    _argumentValues,
		longOptionDescriptors:             make(map[string]Option),
		nextArgumentValueIndex:            uint(0),
		Parsed:                            nil,
		shortOptionDescriptors:            make(map[string]Option),
		skipArgumentsOnNextParseIteration: uint(0),
	}
	return &options
}

// AddOption adds an option to the Options instance based on a provided descriptor.
func (options *Options) AddOption(descriptor Option) {
	switch descriptor.Type {
	case OptionTypeLong:
		options.longOptionDescriptors[descriptor.Name] = descriptor
	case OptionTypeShort:
		options.shortOptionDescriptors[descriptor.Name] = descriptor
	}
}

// AddOptions adds many options to the Options instance at once based on provided descriptors.
func (options *Options) AddOptions(descriptors []Option) {
	for _, descriptor := range descriptors {
		options.AddOption(descriptor)
	}
}

func (options *Options) incrementNextArgumentValueIndex() {
	options.nextArgumentValueIndex++
}

// Parse runs the options parser.
func (options *Options) Parse() *ParseResult {
	if options.Parsed != nil {
		return options.Parsed
	}
	results := make([]ParsedOption, 0)
	for _, argumentValue := range options.argumentValues {
		argumentValue = []rune(strings.TrimSpace(string(argumentValue)))
		// NOTE(@jonathanmarvens): This check is likely unnecessary, but I'm leaving it just in case.
		if string(argumentValue) == "" {
			options.incrementNextArgumentValueIndex()
			continue
		}
		if options.skipArgumentsOnNextParseIteration != uint(0) {
			options.incrementNextArgumentValueIndex()
			options.skipArgumentsOnNextParseIteration--
			continue
		}
		if runtime.GOOS == "windows" &&
			argumentValue[0] == '/' {
			if string(argumentValue) == "/" {
				break
			}
			optionName := argumentValue[1:]
			if utf8.RuneCountInString(string(optionName)) == 1 ||
				(utf8.RuneCountInString(string(optionName)) >= 2 &&
					optionName[1] == ':') ||
				!strings.ContainsRune(string(optionName), ':') {
				shortOptionResults := options.parseShortOptions(optionName, true)
				results = append(results, shortOptionResults...)
			} else {
				longOptionResult := options.parseLongOption(optionName, true)
				results = append(results, longOptionResult)
			}
		} else if argumentValue[0] == '-' {
			if string(argumentValue) == "-" {
				break
			}
			if utf8.RuneCountInString(string(argumentValue)) >= 2 &&
				argumentValue[1] == '-' {
				if string(argumentValue) == "--" {
					options.incrementNextArgumentValueIndex()
					break
				}
				optionName := argumentValue[2:]
				longOptionResult := options.parseLongOption(optionName, false)
				results = append(results, longOptionResult)
			} else {
				optionName := argumentValue[1:]
				shortOptionResults := options.parseShortOptions(optionName, false)
				results = append(results, shortOptionResults...)
			}
		}
	}
	parsedArguments := ParseResult{
		Options:                 make([]ParsedOption, len(results)),
		RemainingArgumentValues: make([]string, 0),
	}
	copy(parsedArguments.Options, results)
	remainingArgumentValues := options.argumentValues[options.nextArgumentValueIndex:]
	for _, argumentValue := range remainingArgumentValues {
		parsedArguments.RemainingArgumentValues = append(parsedArguments.RemainingArgumentValues, string(argumentValue))
	}
	options.Parsed = &parsedArguments
	return options.Parsed
}

func (options *Options) parseLongOption(optionName []rune, dosPrefix bool) ParsedOption {
	var result ParsedOption
	var optionArgument []rune
	if dosPrefix {
		if optionColonSignIndex := strings.IndexRune(string(optionName), ':'); optionColonSignIndex != -1 {
			if len(string(optionName)) >= (optionColonSignIndex + 1) {
				optionArgumentString := string(optionName)[(optionColonSignIndex + 1):]
				optionArgument = make([]rune, 0)
				for i := uint(0); i < uint(len(optionArgumentString)); {
					optionArgumentRune, optionArgumentRuneWidth := utf8.DecodeRuneInString(optionArgumentString[i:])
					optionArgument = append(optionArgument, optionArgumentRune)
					i += uint(optionArgumentRuneWidth)
				}
			}
			optionName = []rune(string(optionName)[:optionColonSignIndex])
		}
	} else {
		if optionEqualSignIndex := strings.IndexRune(string(optionName), '='); optionEqualSignIndex != -1 {
			if len(string(optionName)) >= (optionEqualSignIndex + 1) {
				optionArgumentString := string(optionName)[(optionEqualSignIndex + 1):]
				optionArgument = make([]rune, 0)
				for i := uint(0); i < uint(len(optionArgumentString)); {
					optionArgumentRune, optionArgumentRuneWidth := utf8.DecodeRuneInString(optionArgumentString[i:])
					optionArgument = append(optionArgument, optionArgumentRune)
					i += uint(optionArgumentRuneWidth)
				}
			}
			optionName = []rune(string(optionName)[:optionEqualSignIndex])
		}
	}
	if _, optionDefined := options.longOptionDescriptors[string(optionName)]; !optionDefined {
		result = ParsedOption{
			Error: true,
			Value: newUnknownOptionParseError(options, OptionTypeLong, optionName),
		}
		options.incrementNextArgumentValueIndex()
		return result
	}
	optionDescriptor := options.longOptionDescriptors[string(optionName)]
	switch optionDescriptor.ArgumentType {
	case OptionArgumentTypeNone:
		if optionArgument != nil {
			result = ParsedOption{
				Error: true,
				Value: newExtraneousOptionArgumentParseError(options, OptionTypeLong, optionName, optionArgument),
			}
			options.incrementNextArgumentValueIndex()
			return result
		}
	case OptionArgumentTypeRequired:
		if optionArgument == nil {
			remainingArgumentValues := options.argumentValues[(options.nextArgumentValueIndex + uint(1)):]
			if len(remainingArgumentValues) == 0 {
				result = ParsedOption{
					Error: true,
					Value: newMissingOptionArgumentParseError(options, OptionTypeLong, optionName),
				}
				options.incrementNextArgumentValueIndex()
				return result
			}
			optionArgument = make([]rune, len(remainingArgumentValues[0]))
			copy(optionArgument, remainingArgumentValues[0])
			options.skipArgumentsOnNextParseIteration++
		}
	}
	result = ParsedOption{
		Error: false,
		Value: ParsedOptionValue{
			Argument:     string(optionArgument),
			ArgumentType: optionDescriptor.ArgumentType,
			Name:         string(optionName),
			Type:         optionDescriptor.Type,
		},
	}
	options.incrementNextArgumentValueIndex()
	return result
}

func (options *Options) parseShortOptions(optionNameRunes []rune, dosPrefix bool) []ParsedOption {
	results := make([]ParsedOption, 0)
	skipArgumentsOnNextLocalParseIteration := uint(0)
	for i := uint(0); i < uint(len(string(optionNameRunes))); {
		optionNameString := string(optionNameRunes)[i:]
		_, optionNameRuneWidth := utf8.DecodeRuneInString(optionNameString)
		i += uint(optionNameRuneWidth)
		if skipArgumentsOnNextLocalParseIteration != uint(0) {
			skipArgumentsOnNextLocalParseIteration--
			continue
		}
		optionName := []rune(optionNameString)[0]
		optionArgument := []rune(optionNameString[1:])
		if dosPrefix {
			if strings.IndexRune(string(optionArgument), ':') == 0 {
				optionArgument = []rune(string(optionArgument)[1:])
				skipArgumentsOnNextLocalParseIteration++
			}
		}
		if string(optionArgument) != "" {
			for range optionArgument {
				skipArgumentsOnNextLocalParseIteration++
			}
		}
		if _, optionDefined := options.shortOptionDescriptors[string([]rune{optionName})]; !optionDefined {
			result := ParsedOption{
				Error: true,
				Value: newUnknownOptionParseError(options, OptionTypeShort, []rune{optionName}),
			}
			options.incrementNextArgumentValueIndex()
			results = append(results, result)
			continue
		}
		optionDescriptor := options.shortOptionDescriptors[string([]rune{optionName})]
		switch optionDescriptor.ArgumentType {
		case OptionArgumentTypeNone:
			if string(optionArgument) != "" {
				result := ParsedOption{
					Error: true,
					Value: newExtraneousOptionArgumentParseError(options, OptionTypeShort, []rune{optionName}, optionArgument),
				}
				options.incrementNextArgumentValueIndex()
				results = append(results, result)
				continue
			}
		case OptionArgumentTypeRequired:
			if string(optionArgument) == "" {
				remainingArgumentValues := options.argumentValues[(options.nextArgumentValueIndex + uint(1)):]
				if len(remainingArgumentValues) == 0 {
					result := ParsedOption{
						Error: true,
						Value: newMissingOptionArgumentParseError(options, OptionTypeShort, []rune{optionName}),
					}
					options.incrementNextArgumentValueIndex()
					results = append(results, result)
					continue
				}
				optionArgument = make([]rune, len(remainingArgumentValues[0]))
				copy(optionArgument, remainingArgumentValues[0])
				options.skipArgumentsOnNextParseIteration++
			}
		}
		result := ParsedOption{
			Error: false,
			Value: ParsedOptionValue{
				Argument:     string(optionArgument),
				ArgumentType: optionDescriptor.ArgumentType,
				Name:         string([]rune{optionName}),
				Type:         optionDescriptor.Type,
			},
		}
		options.incrementNextArgumentValueIndex()
		results = append(results, result)
	}
	return results
}
