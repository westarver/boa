package boa

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/bitfield/script"
	"github.com/westarver/helper"
)

const (
	SectionPat = `^\[[[:print:]]+\]`
	LinePat    = `[[:print:]]+ = [[:print:]]*`
	//StructFieldPat = `Type name alias shorthelp longHelp required exclusive isDefault Help beforeLoadHook afterLoadHook Run`
	//ProgramPat     = `program = [[:print:]]*`
	//PackagePat     = `package = [_?a-zA-Z0-9_]*`
	KeyValuePat = `[_?a-zA-Z0-9_]* = (.)*`
	EmptyStr    = `""`
	ImportPath  = "import \"github.com/westarver/argman\""
)

type designParser struct {
	sectionpat *regexp.Regexp
	linepat    *regexp.Regexp
	itemMap    map[string]cmdLineArg
	text       string
	lines      []string
	errs       []parseError
	line       int
}

func (d *designParser) setError(err parseError) {
	d.errs = append(d.errs, err)
}

func (d *designParser) Errors() string {
	var errs []string
	for _, e := range d.errs {
		errs = append(errs, e.Error())
	}
	return strings.Join(errs, "\n")
}

func (d *designParser) appendArg(item cmdLineArg) {
	d.itemMap[item.name] = item
}

type scanDesignfunc func(*designParser) scanDesignfunc

//─────────────┤ getSliceFromFile ├─────────────

func getSliceFromString(s string) []string {
	dsn, err := script.Echo(s).Slice()
	if err != nil {
		log.Fatalf("could not obtain design data`, %v \n", err)
	}
	return dsn
}

//─────────────┤ doGen ├─────────────

func doFromDesign(design string) map[string]cmdLineArg {
	parser := designParser{
		sectionpat: regexp.MustCompile(SectionPat),
		linepat:    regexp.MustCompile(LinePat),
		itemMap:    map[string]cmdLineArg{},
		text:       "",
		lines:      getSliceFromString(design),
		errs:       []parseError{},
		line:       0,
	}

	parser.text = parser.lines[0]
	scanf := scanSection(&parser)
	for {
		if scanf == nil {
			break
		}
		scanf = scanf(&parser)
	}

	fmt.Fprintln(os.Stderr, parser.Errors())
	return parser.itemMap
}

func scanSection(d *designParser) scanDesignfunc {
	// when entering this function we are expecting to be past the line
	// with the section header
	// get the current section text to assign to command/flag name
	section := d.text
	section = strings.TrimPrefix(strings.TrimSuffix(strings.Trim(section, "\t "), "]"), "[")
	if len(section) == 0 {
		d.setError(newError(NoSectionText, "no name in section brackets"))
		return nil
	}

	stringType := reflect.TypeOf("")

	item := cmdLineArg{}
	item.Type = stringType // initial type is string

	if len(section) > 2 {
		if strings.HasPrefix(section, "--") || strings.HasPrefix(section, "-") {
			section = strings.TrimPrefix(section, "--")
			section = strings.TrimPrefix(section, "-")
			item.isFlag = true
		}
	}
	item.name = `"` + section + `"`
	for i := d.line; i < len(d.lines); i++ {
		d.line = i
		line := strings.Trim(d.lines[i], "\t ")
		if len(line) == 0 {
			continue //blank line
		}
		loc := d.sectionpat.FindStringIndex(line)
		if loc != nil {
			d.text = line[loc[0]:loc[1]]
			d.line++
			d.appendArg(item)
			return scanSection
		}

		sl := strings.Split(line, " = ")
		if len(sl) < 2 {
			d.setError(newError(NoKeyValuePair, "bad line in section %s, not a key-value pair", section))
			return nil
		}

		// key value pairs in a section are interpreted as struct fields
		r := regexp.MustCompile(KeyValuePat)
		if r.MatchString(sl[0] + " = " + sl[1]) {
			matchField(d, &item, sl[0], sl[1])
		}
	}
	// output last one
	d.appendArg(item)
	return nil
}

//─────────────┤ matchField ├─────────────

func matchField(d *designParser, it *cmdLineArg, s0, s1 string) {
	flds := "Type isFlag alias shortHelp longHelp required exclusive isDefault"
	loc := helper.FindStringinString(flds, s0)
	if len(loc) == 0 {
		d.setError(newError(NoStructField, "bad line in section, %s not a struct field", s0))
		return
	}
	intType := reflect.TypeOf(int64(0))
	boolType := reflect.TypeOf(true)
	floatType := reflect.TypeOf(float64(0.0))
	stringType := reflect.TypeOf("")
	intSliceType := reflect.TypeOf([]int64{})
	floatSliceType := reflect.TypeOf([]float64{})
	stringSliceType := reflect.TypeOf([]string{})

	switch s0 {
	case "Type":
		if s1 == "string" {
			it.Type = stringType
			return
		}

		if s1 == "[]string" {
			it.Type = stringSliceType
			return
		}

		if s1 == "bool" {
			it.Type = boolType
			return
		}

		if strings.HasPrefix(s1, "float") {
			it.Type = floatType
			return
		}

		if strings.HasPrefix(s1, "[]float") {
			it.Type = floatSliceType
			return
		}

		if strings.HasPrefix(s1, "int") {
			it.Type = intType
			return
		}

		if strings.HasPrefix(s1, "[]int") {
			it.Type = intSliceType
			return
		}
	case "alias":
		if s1 == EmptyStr {
			it.alias = s1
		} else {
			it.alias = `"` + s1 + `"`
		}
	case "shortHelp":
		if s1 == EmptyStr {
			it.shortHelp = s1
		} else {
			it.shortHelp = `"` + s1 + `"`
		}
	case "longHelp":
		if s1 == EmptyStr {
			it.longHelp = s1
		} else {
			it.longHelp = "`" + s1 + "`"
		}
	case "isFlag":
		if s1 == "true" {
			it.isFlag = true
		} else {
			it.isFlag = false
		}
	case "required":
		if s1 == "true" {
			it.required = true
		} else {
			it.required = false
		}
	case "exclusive":
		if s1 == "true" {
			it.exclusive = true
		} else {
			it.exclusive = false
		}
	case "isDefault":
		if s1 == "true" {
			it.isDefault = true
		} else {
			it.isDefault = false
		}
	// the functions will have to be added at run time through the Set*Fn methods
	default:
		d.setError(newError(NoStructField, "bad line in section, %s not a struct field", s0))
	}
}
