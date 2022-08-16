package boa

import "fmt"

type parseError struct {
	code parseErrCode
	err  error
}

type parseErrCode int

const (
	NoFileGiven parseErrCode = iota
	NoNameGiven
	WrongFileFormat
	FileReadError
	DataIncomplete
	NoStructField
	NoKeyValuePair
	NoSectionText
	UnsupportedType
)

func (e parseError) Error() string {
	return fmt.Sprintf("error parsing design data with code %d: %s", e.code, e.err.Error())
}

func newError(code parseErrCode, fmtstr string, args ...any) parseError {
	return parseError{
		code: code,
		err:  fmt.Errorf(fmtstr, args...),
	}
}
