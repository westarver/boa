package boa

import "fmt"

type ParseError struct {
	Code ParseErrCode
	Err  error
}

type ParseErrCode int

func (c ParseErrCode) String() string {
	return codestr(c)
}

const (
	//errors from reading input script
	BeExternalError ParseErrCode = iota
	//"expecting a file path for input script"
	BeNoFileGiven
	//"input script is not formatted as a valid Boa json input"
	BeWrongFileFormat
	//"cannot read input script %s"
	BeFileReadError
	// "unsupported argument type"
	BeUnsupportedType
	//"reached end of input script"
	BeEofError
	//"line only contains meta characters"
	BeBadMetaLine
	//"meta character string is not at beginning of line"
	BeMetaNotStart
	//"command or flag in input script cannot be  parsed"
	BeNoCommandName

	//errors from parsing command/flag arguments

	//"item %s is exclusive but was found with %s"
	BeNoExclusiveItem
	//"item %s is required but was not found"
	BeNoRequiredItem
	//"command or flag passed on command line is not recognized"
	BeInvalidCommand
	//"argument for %s not found"
	BeNoRequiredString
	//"integer argument for %s not found"
	BeNoRequiredInt
	//"real number argument for %s not found"
	BeNoRequiredFloat
	//"date argument for %s not found"
	BeNoRequiredDate
	// "time argument for %s not found"
	BeNoRequiredTime
	//"time duration argument for %s not found"
	BeNoRequiredDuration
	//"file path argument for %s not found"
	BeNoRequiredPath
	//"URL argument for %s not found"
	BeNoRequiredURL
	//"Email address argument for %s not found"
	BeNoRequiredEmail
	//"Phone number argument for %s not found"
	BeNoRequiredPhone
	//"IP address argument for %s not found"
	BeNoRequiredIPv4
	//"%s, argument for %s, cannot be interpreted as a boolean"
	BeNotABool
	//"%s, argument for %s, cannot be interpreted as an integer"
	BeNotAnInt
	// "%s, argument for %s, cannot be interpreted as a real number"
	BeNotAFloat
	// "argument to %s, %s is not a valid date value such as '01-01-2022'"
	BeNotADate
	//"argument to %s, %s is not a valid time value such as '3:45PM'"
	BeNotATime
	//"argument to %s, %s is not a valid duration value such as '1h10m20s'"
	BeNotADuration
	//"%s, argument for %s, cannot be interpreted as an email address"
	BeNotAnEmail
	//"%s, argument for %s, cannot be interpreted as a phone number"
	BeNotAPhone
	//"%s, argument for %s, cannot be interpreted as a file path"
	BeNotAPath
	//"%s, argument for %s, cannot be interpreted as a URL"
	BeNotAURL
	//"%s, argument for %s, cannot be interpreted as an IP address of IPv4 format"
	BeNotAnIPv4
)

func (c ParseErrCode) fmts() string {
	return stringFromCode(c)
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}

func newParseError(code ParseErrCode, fmtstr string, args ...any) ParseError {
	return ParseError{
		Code: code,
		Err:  fmt.Errorf(fmtstr, args...),
	}
}

func Errorf(code ParseErrCode, args ...any) ParseError {
	return newParseError(code, code.fmts(), args...)
}

func stringFromCode(code ParseErrCode) string {
	switch code {
	//errors from reading input script
	case BeNoFileGiven:
		return "expecting a file path for input"
	case BeWrongFileFormat:
		return "input script is not formatted as a Boa input"
	case BeFileReadError:
		return "cannot read input %s"
	case BeUnsupportedType:
		return "unsupported argument type"
	case BeEofError:
		return "unexpected end of input"
	case BeNoCommandName:
		return "command or flag in input cannot be  parsed"

		// errors from parsing command/flag arguments

	case BeNoExclusiveItem:
		return "item %s is exclusive but was found with %s"
	case BeNoRequiredItem:
		return "Item %s is required but was not found"
	case BeInvalidCommand:
		return "%s: command or flag passed on command line is not recognized"
	case BeNoRequiredString:
		return "argument for %s not found"
	case BeNoRequiredInt:
		return "integer argument for %s not found"
	case BeNoRequiredFloat:
		return "real number argument for %s not found"
	case BeNoRequiredDate:
		return "argument to %s, %s is not a valid date value such as 'mm-dd-yyyy'"
	case BeNoRequiredTime:
		return "time argument for %s not found"
	case BeNoRequiredDuration:
		return "time duration argument for %s not found"
	case BeNoRequiredPath:
		return "file path argument for %s not found"
	case BeNoRequiredURL:
		return "URL argument for %s not found"
	case BeNoRequiredEmail:
		return "Email address argument for %s not found"
	case BeNoRequiredPhone:
		return "Phone number argument for %s not found"
	case BeNoRequiredIPv4:
		return "IP address argument for %s not found"
	case BeNotABool:
		return "%s, argument for %s, cannot be interpreted as a boolean"
	case BeNotAnInt:
		return "%s, argument for %s, cannot be interpreted as an integer"
	case BeNotAFloat:
		return "%s, argument for %s, cannot be interpreted as a real number"
	case BeNotADate:
		return "%s, argument for %s, cannot be interpreted as a date"
	case BeNotATime:
		return "argument to %s, %s is not a valid time value such as '3:45PM'"
	case BeNotADuration:
		return "argument to %s, %s is not a valid duration value such as '1h10m20s'"
	case BeNotAnEmail:
		return "%s, argument for %s, cannot be interpreted as an email address"
	case BeNotAPhone:
		return "%s, argument for %s, cannot be interpreted as a phone number"
	case BeNotAPath:
		return "%s, argument for %s, cannot be interpreted as a file path"
	case BeNotAURL:
		return "%s, argument for %s, cannot be interpreted as a URL"
	case BeNotAnIPv4:
		return "%s, argument for %s, cannot be interpreted as an IP address of IPv4 format"
	}
	return "Unknown error"
}

func codestr(code ParseErrCode) string {
	switch code {
	//errors from reading input script
	case BeNoFileGiven:
		return "NoFileGiven"
	case BeWrongFileFormat:
		return "WrongFileFormat"
	case BeFileReadError:
		return "FileReadError"
	case BeUnsupportedType:
		return "UnsupportedType"
	case BeEofError:
		return "EofError"
	case BeNoCommandName:
		return "NoCommandName"

		// errors from parsing command/flag arguments

	case BeNoExclusiveItem:
		return "NoExclusiveItem"
	case BeNoRequiredItem:
		return "NoRequiredItem"
	case BeInvalidCommand:
		return "InvalidCommand"
	case BeNoRequiredString:
		return "NoRequiredString"
	case BeNoRequiredInt:
		return "NoRequiredInt"
	case BeNoRequiredFloat:
		return "NoRequiredFloat"
	case BeNoRequiredDate:
		return "NoRequiredDate"
	case BeNoRequiredTime:
		return "NoRequiredTime"
	case BeNoRequiredDuration:
		return "NoRequiredDuration"
	case BeNoRequiredPath:
		return "NoRequiredPath"
	case BeNoRequiredURL:
		return "NoRequiredURL"
	case BeNoRequiredEmail:
		return "NoRequiredEmail"
	case BeNoRequiredPhone:
		return "NoRequiredPhone"
	case BeNoRequiredIPv4:
		return "NoRequiredIPv4"
	case BeNotABool:
		return "NotABool"
	case BeNotAnInt:
		return "NotAnInt"
	case BeNotAFloat:
		return "NotAFloat"
	case BeNotADate:
		return "NotADate"
	case BeNotATime:
		return "BeNotATime"
	case BeNotADuration:
		return "NotADuration"
	case BeNotAnEmail:
		return "NotAnEmail"
	case BeNotAPhone:
		return "NotAPhone"
	case BeNotAPath:
		return "NotAPath"
	case BeNotAURL:
		return "NotAURL"
	case BeNotAnIPv4:
		return "NotAnIPv4"
	}
	return "Unknown error code"
}
