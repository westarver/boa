package boa

import (
	"strings"
)

type CLI struct {
	Items   map[string]CmdLineItem
	AllHelp map[string]string
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

func (C *CLI) HasErrors() bool {
	return len(C.errs) != 0
}

func (C *CLI) LastError() error {
	if C.HasErrors() {
		return C.errs[len(C.errs)-1]
	}
	return nil
}

func (C *CLI) Bool(item string) (bool, bool) {
	if b, ok := C.Items[item].value.(bool); ok {
		return b, ok
	}
	return false, false
}

func (C *CLI) String(item string) (string, bool) {
	if s, ok := C.Items[item].value.(string); ok {
		return s, ok
	}
	return "", false
}

func (C *CLI) Int(item string) (int, bool) {
	if n, ok := C.Items[item].value.(int); ok {
		return n, ok
	}
	return 0, false
}

func (C *CLI) Float(item string) (float64, bool) {
	if f, ok := C.Items[item].value.(float64); ok {
		return f, ok
	}
	return 0.0, false
}

func (C *CLI) StringSlice(item string) ([]string, bool) {
	if s, ok := C.Items[item].value.([]string); ok {
		return s, ok
	}
	return nil, false
}

func (C *CLI) IntSlice(item string) ([]int, bool) {
	if s, ok := C.Items[item].value.([]int); ok {
		return s, ok
	}
	return nil, false
}

func (C *CLI) FloatSlice(item string) ([]float64, bool) {
	if s, ok := C.Items[item].value.([]float64); ok {
		return s, ok
	}
	return nil, false
}

type HelpType int

const (
	ShortStr HelpType = iota
	LongStr
	CombinedStr
)

func (C *CLI) Help(topic string, ty ...HelpType) string {
	typ := CombinedStr

	if len(ty) > 0 {
		typ = ty[0]
	}

	// deal with alias passed in
	for _, c := range C.Items {
		if c.alias == topic {
			topic = c.name
			break
		}
	}

	switch typ {
	case ShortStr:
		if i, ok := C.Items[topic]; ok {
			return i.ShortHelp()
		}
	case LongStr:
		if i, ok := C.Items[topic]; ok {
			return i.LongHelp()
		}
	case CombinedStr:
		return C.AllHelp[topic]
	}

	return ""
}

type CmdLineItem struct {
<<<<<<< HEAD
	paramType   ParamType
=======
	Type        ValueType
>>>>>>> ddb2b57a0cb42366fc393fe1a983ecea453ad4b8
	value       any
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
	id          int // use as index to order the commands as read
	paramOpt    bool
	paramCount  int
}

// (CmdLineItem) Name returns the value of the unexported struct field name
func (C CmdLineItem) Name() string {
	return C.name
}

// (CmdLineItem) Alias returns the value of the unexported struct field alias
func (C CmdLineItem) Alias() string {
	return C.alias
}

// (CmdLineItem) ShortHelp returns the value of the unexported struct field shortHelp
func (C CmdLineItem) ShortHelp() string {
	return C.shortHelp
}

// (CmdLineItem) LongHelp returns the value of the unexported struct field longHelp
func (C CmdLineItem) LongHelp() string {
	return C.longHelp
}

// (CmdLineItem) IsDefault returns the value of the unexported struct field isDefault
func (C CmdLineItem) IsDefault() bool {
	return C.isDefault
}

// (CmdLineItem) IsFlag returns the value of the unexported struct field isFlag
func (C CmdLineItem) IsFlag() bool {
	return C.isFlag
}

// (CmdLineItem) Exclusive returns the value of the unexported struct field exclusive
func (C CmdLineItem) Exclusive() bool {
	return C.exclusive
}

// (CmdLineItem) Required returns the value of the unexported struct field requireStatus
func (C CmdLineItem) Required() bool {
	return C.required
}

// (CmdLineItem) RequiredOr returns the value of the unexported struct field requiredOr
func (C CmdLineItem) RequiredOr() []string {
	return C.requiredOr
}

// (CmdLineItem) RequiredAnd returns the value of the unexported struct field requiredAnd
func (C CmdLineItem) RequiredAnd() []string {
	return C.requiredAnd
}

<<<<<<< HEAD
func (C CmdLineItem) ID() int {
	return C.id
}

func (C CmdLineItem) ParamCount() int {
	return C.paramCount
}

func (C CmdLineItem) ParamOpt() bool {
	return C.paramOpt
}

func (C CmdLineItem) ParamType() ParamType {
	return C.paramType
}

// (*CmdLineItem.Value()returns the value of the unexported struct field value
// this function will require a concrete type assertion.
// call using strval := obj.Value().(string)
func (C CmdLineItem) Value() any {
	return C.value
}
=======
// (*CmdLineItem.Value()returns the value of the unexported struct field value
// this function will require a concrete type assertion.
// call using strval := obj.Value().(string)
func (C CmdLineItem) Value() any {
	return C.value
}
>>>>>>> ddb2b57a0cb42366fc393fe1a983ecea453ad4b8
