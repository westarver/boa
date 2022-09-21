package boa

import (
	"net"
	"net/mail"
	"net/url"
	"strings"
	"time"
)

type CLI struct {
	Application string
	Items       map[string]CmdLineItem
	AllHelp     map[string]string
	Errs        []error
}

func (C *CLI) Errors() string {
	var errs []string
	for _, e := range C.Errs {
		errs = append(errs, e.Error())
	}
	return strings.Join(errs, "\n")
}

func (C *CLI) SetError(err error) {
	C.Errs = append(C.Errs, err)
}

func (C *CLI) HasErrors() bool {
	return len(C.Errs) != 0
}

func (C *CLI) LastError() error {
	if C.HasErrors() {
		return C.Errs[len(C.Errs)-1]
	}
	return nil
}

func (C *CLI) Bool(item string) (bool, bool) {
	if b, ok := C.Items[item].Value.(bool); ok {
		return b, ok
	}
	return false, false
}

func (C *CLI) String(item string) (string, bool) {
	if s, ok := C.Items[item].Value.(string); ok {
		return s, ok
	}
	return "", false
}

func (C *CLI) Int(item string) (int, bool) {
	if n, ok := C.Items[item].Value.(int); ok {
		return n, ok
	}
	return 0, false
}

func (C *CLI) Float(item string) (float64, bool) {
	if f, ok := C.Items[item].Value.(float64); ok {
		return f, ok
	}
	return 0.0, false
}

func (C *CLI) StringSlice(item string) ([]string, bool) {
	if s, ok := C.Items[item].Value.([]string); ok {
		return s, ok
	}
	return nil, false
}

func (C *CLI) IntSlice(item string) ([]int, bool) {
	if n, ok := C.Items[item].Value.([]int); ok {
		return n, ok
	}
	return nil, false
}

func (C *CLI) FloatSlice(item string) ([]float64, bool) {
	if f, ok := C.Items[item].Value.([]float64); ok {
		return f, ok
	}
	return nil, false
}

func (C *CLI) Time(item string) (time.Time, bool) {
	if t, ok := C.Items[item].Value.(time.Time); ok {
		return t, ok
	}
	return time.Time{}, false
}

func (C *CLI) TimeSSlice(item string) ([]time.Time, bool) {
	if t, ok := C.Items[item].Value.([]time.Time); ok {
		return t, ok
	}
	return nil, false
}

func (C *CLI) TimeDuration(item string) (time.Duration, bool) {
	if t, ok := C.Items[item].Value.(time.Duration); ok {
		return t, ok
	}
	return time.Duration(0), false
}

func (C *CLI) TimeDurationSlice(item string) ([]time.Duration, bool) {
	if t, ok := C.Items[item].Value.([]time.Duration); ok {
		return t, ok
	}
	return nil, false
}

func (C *CLI) Date(item string) (time.Time, bool) {
	if d, ok := C.Items[item].Value.(time.Time); ok {
		return d, ok
	}
	return time.Time{}, false
}

func (C *CLI) DateSlice(item string) ([]time.Time, bool) {
	if d, ok := C.Items[item].Value.([]time.Time); ok {
		return d, ok
	}
	return nil, false
}

func (C *CLI) Path(item string) (string, bool) {
	if p, ok := C.Items[item].Value.(string); ok {
		return p, ok
	}
	return "", false
}

func (C *CLI) PathSlice(item string) ([]string, bool) {
	if p, ok := C.Items[item].Value.([]string); ok {
		return p, ok
	}
	return nil, false
}

func (C *CLI) Email(item string) (mail.Address, bool) {
	if m, ok := C.Items[item].Value.(mail.Address); ok {
		return m, ok
	}
	return mail.Address{}, false
}

func (C *CLI) EmailSlice(item string) ([]mail.Address, bool) {
	if m, ok := C.Items[item].Value.([]mail.Address); ok {
		return m, ok
	}
	return nil, false
}

func (C *CLI) IPv4(item string) (net.IP, bool) {
	if ip, ok := C.Items[item].Value.(net.IP); ok {
		return ip, ok
	}
	return net.IP{}, false
}

func (C *CLI) IPv4Slice(item string) ([]net.IP, bool) {
	if ip, ok := C.Items[item].Value.([]net.IP); ok {
		return ip, ok
	}
	return nil, false
}

func (C *CLI) URL(item string) (url.URL, bool) {
	if u, ok := C.Items[item].Value.(url.URL); ok {
		return u, ok
	}
	return url.URL{}, false
}

func (C *CLI) URLSlice(item string) ([]url.URL, bool) {
	if u, ok := C.Items[item].Value.([]url.URL); ok {
		return u, ok
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
		if c.Alias == topic {
			topic = c.Name
			break
		}
	}

	switch typ {
	case ShortStr:
		if i, ok := C.Items[topic]; ok {
			return i.ShortHelp
		}
	case LongStr:
		if i, ok := C.Items[topic]; ok {
			return i.LongHelp
		}
	case CombinedStr:
		return C.AllHelp[topic]
	}

	return ""
}

type CmdLineItem struct {
	ParamType   ParameterType
	Value       any
	Name        string
	Alias       string
	ShortHelp   string
	LongHelp    string
	IsDefault   bool
	IsFlag      bool
	Exclusive   bool
	Required    bool
	RequiredOr  []string
	RequiredAnd []string
	ParamOpt    bool
	ParamCount  int
	Extra       string // the boa-gui tool uses this field for code generation
	Disabled    bool   // the boa-gui tool uses this field for script editing
	Id          int    // use as index to sort the items in as read order
}

// // (CmdLineItem) Name returns the value of the unexported struct field name
// func (C CmdLineItem) Name() string {
// 	return C.Name
// }

// // (CmdLineItem) Alias returns the value of the unexported struct field alias
// func (C CmdLineItem) Alias() string {
// 	return C.alias
// }

// // (CmdLineItem) ShortHelp returns the value of the unexported struct field shortHelp
// func (C CmdLineItem) ShortHelp() string {
// 	return C.shortHelp
// }

// // (CmdLineItem) LongHelp returns the value of the unexported struct field longHelp
// func (C CmdLineItem) LongHelp() string {
// 	return C.longHelp
// }

// // (CmdLineItem) IsDefault returns the value of the unexported struct field isDefault
// func (C CmdLineItem) IsDefault() bool {
// 	return C.isDefault
// }

// // (CmdLineItem) IsFlag returns the value of the unexported struct field isFlag
// func (C CmdLineItem) IsFlag() bool {
// 	return C.isFlag
// }

// // (CmdLineItem) Exclusive returns the value of the unexported struct field exclusive
// func (C CmdLineItem) Exclusive() bool {
// 	return C.exclusive
// }

// // (CmdLineItem) Required returns the value of the unexported struct field requireStatus
// func (C CmdLineItem) Required() bool {
// 	return C.required
// }

// // (CmdLineItem) RequiredOr returns the value of the unexported struct field requiredOr
// func (C CmdLineItem) RequiredOr() []string {
// 	return C.requiredOr
// }

// // (CmdLineItem) RequiredAnd returns the value of the unexported struct field requiredAnd
// func (C CmdLineItem) RequiredAnd() []string {
// 	return C.requiredAnd
// }

// func (C CmdLineItem) ID() int {
// 	return C.id
// }

// func (C CmdLineItem) ParamCount() int {
// 	return C.paramCount
// }

// func (C CmdLineItem) ParamOpt() bool {
// 	return C.paramOpt
// }

// func (C CmdLineItem) ParamType() ParameterType {
// 	return C.paramType
// }

// func (C CmdLineItem) Extra() string {
// 	return C.extra
// }

// func (C CmdLineItem) Disabled() bool {
// 	return C.disabled
// }

// // (*CmdLineItem.Value()returns the value of the unexported struct field value
// // this function will require a concrete type assertion.
// // call using strval := obj.Value().(string)
// func (C CmdLineItem) Value() any {
// 	return C.value
// }

// // (*CmdLineItem) SetparamType assigns the value of val to the unexported struct field paramType
// func (C *CmdLineItem) SetparamType(val ParameterType) {
// 	C.paramType = val
// }

// // (*CmdLineItem) Setname assigns the value of val to the unexported struct field name
// func (C *CmdLineItem) Setname(val string) {
// 	C.name = val
// }

// // (*CmdLineItem) Setalias assigns the value of val to the unexported struct field alias
// func (C *CmdLineItem) Setalias(val string) {
// 	C.alias = val
// }

// // (*CmdLineItem) SetshortHelp assigns the value of val to the unexported struct field shortHelp
// func (C *CmdLineItem) SetshortHelp(val string) {
// 	C.shortHelp = val
// }

// // (*CmdLineItem) SetlongHelp assigns the value of val to the unexported struct field longHelp
// func (C *CmdLineItem) SetlongHelp(val string) {
// 	C.longHelp = val
// }

// // (*CmdLineItem) SetisDefault assigns the value of val to the unexported struct field isDefault
// func (C *CmdLineItem) SetisDefault(val bool) {
// 	C.isDefault = val
// }

// // (*CmdLineItem) SetisFlag assigns the value of val to the unexported struct field isFlag
// func (C *CmdLineItem) SetisFlag(val bool) {
// 	C.isFlag = val
// }

// // (*CmdLineItem) Setexclusive assigns the value of val to the unexported struct field exclusive
// func (C *CmdLineItem) Setexclusive(val bool) {
// 	C.exclusive = val
// }

// // (*CmdLineItem) Setrequired assigns the value of val to the unexported struct field required
// func (C *CmdLineItem) Setrequired(val bool) {
// 	C.required = val
// }

// // (CmdLineItem) RequiredOr assigns the value of val to the unexported struct field requiredOr
// func (C *CmdLineItem) SetrequiredOr(val []string) {
// 	C.requiredOr = val
// }

// // (CmdLineItem) RequiredAnd assigns the value of val to the unexported struct field requiredAnd
// func (C *CmdLineItem) SetrequiredAnd(val []string) {
// 	C.requiredAnd = val
// }

// // (*CmdLineItem) Setid assigns the value of val to the unexported struct field id
// func (C *CmdLineItem) Setid(val int) {
// 	C.id = val
// }

// // (*CmdLineItem) SetparamOpt assigns the value of val to the unexported struct field paramOpt
// func (C *CmdLineItem) SetparamOpt(val bool) {
// 	C.paramOpt = val
// }

// // (*CmdLineItem) SetparamCount assigns the value of val to the unexported struct field paramCount
// func (C *CmdLineItem) SetparamCount(val int) {
// 	C.paramCount = val
// }

// func (C *CmdLineItem) Setextra(val string) {
// 	C.extra = val
// }

// func (C *CmdLineItem) Setdisabled(val bool) {
// 	C.disabled = val
// }
