package boa

import (
	"reflect"
	"strings"
)

type CmdLineItem[T any] struct {
	value       T
	name        string
	alias       string
	shortHelp   string
	longHelp    string
	isDefault   bool
	isFlag      bool
	exclusive   bool
	required    bool
	requiredOr  []string
	requiredAnd []string
	// Helpfn       func(io.Writer)
	// PostLoadHook func(*CLI) error
	// Runfn        func()
}

type GenericArg interface {
	Name() string
	Alias() string
	ShortHelp() string
	LongHelp() string
	IsDefault() bool
	IsFlag() bool
	Exclusive() bool
	Required() bool
	RequiredOr() []string
	RequiredAnd() []string
}

type CLI struct {
	Items   map[string]any
	AllHelp map[string]string
	Post    func(*CLI)
	errs    []error
}

func (C *CLI) Errors() string {
	var errs []string
	for _, e := range C.errs {
		errs = append(errs, e.Error())
	}
	return strings.Join(errs, "\n")
}

func (C *CLI) SetError(err error) {
	C.errs = append(C.errs, err)
}

type cmdLineArg struct {
	Type        reflect.Type
	name        string
	alias       string
	shortHelp   string
	longHelp    string
	isDefault   bool
	isFlag      bool
	exclusive   bool
	required    bool
	canHaveGlob bool
	requiredOr  []string
	requiredAnd []string
}

// (*CmdLineItem[T]) Name returns the value of the unexported struct field name
func (C CmdLineItem[T]) Name() string {
	return C.name
}

// (*CmdLineItem[T]) Alias returns the value of the unexported struct field alias
func (C CmdLineItem[T]) Alias() string {
	return C.alias
}

// (*CmdLineItem[T]) ShortHelp returns the value of the unexported struct field shortHelp
func (C CmdLineItem[T]) ShortHelp() string {
	return C.shortHelp
}

// (*CmdLineItem[T]) LongHelp returns the value of the unexported struct field longHelp
func (C CmdLineItem[T]) LongHelp() string {
	return C.shortHelp
}

// (*CmdLineItem[T]) IsDefault returns the value of the unexported struct field isDefault
func (C CmdLineItem[T]) IsDefault() bool {
	return C.isDefault
}

// (*CmdLineItem[T]) IsFlag returns the value of the unexported struct field isFlag
func (C CmdLineItem[T]) IsFlag() bool {
	return C.isFlag
}

// (*CmdLineItem[T]) Exclusive returns the value of the unexported struct field exclusive
func (C CmdLineItem[T]) Exclusive() bool {
	return C.exclusive
}

// (*CmdLineItem[T]) Required returns the value of the unexported struct field requireStatus
func (C CmdLineItem[T]) Required() bool {
	return C.required
}

// (*CmdLineItem[T]) RequiredOr returns the value of the unexported struct field requiredOr
func (C CmdLineItem[T]) RequiredOr() []string {
	return C.requiredOr
}

// (*CmdLineItem[T]) requiredAnd returns the value of the unexported struct field requiredAnd
func (C CmdLineItem[T]) RequiredAnd() []string {
	return C.requiredAnd
}

// (*CmdLineItem[T]).Value()returns the value of the unexported struct field value
// this function will require a concrete type assertion.
// call using strval := obj.(CmdLineItem[<string>]).Value()
func (C *CmdLineItem[T]) Value() T {
	return C.value
}
