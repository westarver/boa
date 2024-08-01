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
	// both errors accum in CLI and the errors of each CmdLineItem
	// are reported with this func
	var errs []string
	for _, c := range C.Items {
		errs = append(errs, c.Error())
	}
	for _, e := range C.Errs {
		errs = append(errs, e.Error())
	}
	return strings.Join(errs, "\n")
}

func (C *CLI) SetError(err error) {
	C.Errs = append(C.Errs, err)
}

func (C *CLI) HasErrors() bool {
	return len(C.Errors()) != 0
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
	Id           int // use as index to sort the items in as read order
	Name         string
	Alias        string
	ParamType    ParameterType
	ParamCount   int // -100 means 1 or more are required, -99 is 0 or more
	ShortHelp    string
	LongHelp     string
	Errors       []error
	Value        interface{} // string values taken from command line may be converted to any type
	DefaultValue string      // string because all values are taken off the command line as strings

	IsDefault   bool
	IsFlag      bool
	IsExclusive bool
	IsParamOpt  bool
	IsRequired  bool
	IsDeleted   bool

	RunCode string // the boa-gui tool uses this field for code generation
	ParName string
	ChNames []string
}

func (c CmdLineItem) Error() string {
	var er []string
	for _, e := range c.Errors {
		er = append(er, e.Error())
	}
	return strings.Join(er, "\n")
}

func (c *CmdLineItem) SetError(err error) {
	c.Errors = append(c.Errors, err)
}

func NewCmdLineItem() *CmdLineItem {
	return &CmdLineItem{ChNames: nil}
}
