// Package json parses JSON as defined by [1].
//
// BUGS: the escaped forward solidus (`\/`) is not currently handled.
//
// [1]: http://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf
package json

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

var g = &grammar{
	rules: []*rule{
		{
			name: "JSON",
			pos:  position{line: 17, col: 1, offset: 363},
			expr: &actionExpr{
				pos: position{line: 17, col: 8, offset: 372},
				run: (*parser).callonJSON1,
				expr: &seqExpr{
					pos: position{line: 17, col: 8, offset: 372},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 17, col: 8, offset: 372},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 17, col: 10, offset: 374},
							label: "vals",
							expr: &oneOrMoreExpr{
								pos: position{line: 17, col: 15, offset: 379},
								expr: &ruleRefExpr{
									pos:  position{line: 17, col: 15, offset: 379},
									name: "Value",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 17, col: 22, offset: 386},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "Value",
			pos:  position{line: 29, col: 1, offset: 589},
			expr: &actionExpr{
				pos: position{line: 29, col: 9, offset: 599},
				run: (*parser).callonValue1,
				expr: &seqExpr{
					pos: position{line: 29, col: 9, offset: 599},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 29, col: 9, offset: 599},
							label: "val",
							expr: &choiceExpr{
								pos: position{line: 29, col: 15, offset: 605},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 29, col: 15, offset: 605},
										name: "Object",
									},
									&ruleRefExpr{
										pos:  position{line: 29, col: 24, offset: 614},
										name: "Array",
									},
									&ruleRefExpr{
										pos:  position{line: 29, col: 32, offset: 622},
										name: "Number",
									},
									&ruleRefExpr{
										pos:  position{line: 29, col: 41, offset: 631},
										name: "String",
									},
									&ruleRefExpr{
										pos:  position{line: 29, col: 50, offset: 640},
										name: "Bool",
									},
									&ruleRefExpr{
										pos:  position{line: 29, col: 57, offset: 647},
										name: "Null",
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 29, col: 64, offset: 654},
							name: "_",
						},
					},
				},
			},
		},
		{
			name: "Object",
			pos:  position{line: 33, col: 1, offset: 685},
			expr: &actionExpr{
				pos: position{line: 33, col: 10, offset: 696},
				run: (*parser).callonObject1,
				expr: &seqExpr{
					pos: position{line: 33, col: 10, offset: 696},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 33, col: 10, offset: 696},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 33, col: 14, offset: 700},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 33, col: 16, offset: 702},
							label: "vals",
							expr: &zeroOrOneExpr{
								pos: position{line: 33, col: 21, offset: 707},
								expr: &seqExpr{
									pos: position{line: 33, col: 23, offset: 709},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 33, col: 23, offset: 709},
											name: "String",
										},
										&ruleRefExpr{
											pos:  position{line: 33, col: 30, offset: 716},
											name: "_",
										},
										&litMatcher{
											pos:        position{line: 33, col: 32, offset: 718},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 33, col: 36, offset: 722},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 33, col: 38, offset: 724},
											name: "Value",
										},
										&zeroOrMoreExpr{
											pos: position{line: 33, col: 44, offset: 730},
											expr: &seqExpr{
												pos: position{line: 33, col: 46, offset: 732},
												exprs: []interface{}{
													&litMatcher{
														pos:        position{line: 33, col: 46, offset: 732},
														val:        ",",
														ignoreCase: false,
													},
													&ruleRefExpr{
														pos:  position{line: 33, col: 50, offset: 736},
														name: "_",
													},
													&ruleRefExpr{
														pos:  position{line: 33, col: 52, offset: 738},
														name: "String",
													},
													&ruleRefExpr{
														pos:  position{line: 33, col: 59, offset: 745},
														name: "_",
													},
													&litMatcher{
														pos:        position{line: 33, col: 61, offset: 747},
														val:        ":",
														ignoreCase: false,
													},
													&ruleRefExpr{
														pos:  position{line: 33, col: 65, offset: 751},
														name: "_",
													},
													&ruleRefExpr{
														pos:  position{line: 33, col: 67, offset: 753},
														name: "Value",
													},
												},
											},
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 33, col: 79, offset: 765},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Array",
			pos:  position{line: 48, col: 1, offset: 1122},
			expr: &actionExpr{
				pos: position{line: 48, col: 9, offset: 1132},
				run: (*parser).callonArray1,
				expr: &seqExpr{
					pos: position{line: 48, col: 9, offset: 1132},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 48, col: 9, offset: 1132},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 48, col: 13, offset: 1136},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 48, col: 15, offset: 1138},
							label: "vals",
							expr: &zeroOrOneExpr{
								pos: position{line: 48, col: 20, offset: 1143},
								expr: &seqExpr{
									pos: position{line: 48, col: 22, offset: 1145},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 48, col: 22, offset: 1145},
											name: "Value",
										},
										&zeroOrMoreExpr{
											pos: position{line: 48, col: 28, offset: 1151},
											expr: &seqExpr{
												pos: position{line: 48, col: 30, offset: 1153},
												exprs: []interface{}{
													&litMatcher{
														pos:        position{line: 48, col: 30, offset: 1153},
														val:        ",",
														ignoreCase: false,
													},
													&ruleRefExpr{
														pos:  position{line: 48, col: 34, offset: 1157},
														name: "_",
													},
													&ruleRefExpr{
														pos:  position{line: 48, col: 36, offset: 1159},
														name: "Value",
													},
												},
											},
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 48, col: 48, offset: 1171},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 62, col: 1, offset: 1491},
			expr: &actionExpr{
				pos: position{line: 62, col: 10, offset: 1502},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 62, col: 10, offset: 1502},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 62, col: 10, offset: 1502},
							expr: &litMatcher{
								pos:        position{line: 62, col: 10, offset: 1502},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 62, col: 15, offset: 1507},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 62, col: 23, offset: 1515},
							expr: &seqExpr{
								pos: position{line: 62, col: 25, offset: 1517},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 62, col: 25, offset: 1517},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 62, col: 29, offset: 1521},
										expr: &ruleRefExpr{
											pos:  position{line: 62, col: 29, offset: 1521},
											name: "DecimalDigit",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 62, col: 46, offset: 1538},
							expr: &ruleRefExpr{
								pos:  position{line: 62, col: 46, offset: 1538},
								name: "Exponent",
							},
						},
					},
				},
			},
		},
		{
			name: "Integer",
			pos:  position{line: 68, col: 1, offset: 1699},
			expr: &choiceExpr{
				pos: position{line: 68, col: 11, offset: 1711},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 68, col: 11, offset: 1711},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 68, col: 17, offset: 1717},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 68, col: 17, offset: 1717},
								name: "NonZeroDecimalDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 68, col: 37, offset: 1737},
								expr: &ruleRefExpr{
									pos:  position{line: 68, col: 37, offset: 1737},
									name: "DecimalDigit",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Exponent",
			pos:  position{line: 70, col: 1, offset: 1754},
			expr: &seqExpr{
				pos: position{line: 70, col: 12, offset: 1767},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 70, col: 12, offset: 1767},
						val:        "e",
						ignoreCase: true,
					},
					&zeroOrOneExpr{
						pos: position{line: 70, col: 17, offset: 1772},
						expr: &charClassMatcher{
							pos:        position{line: 70, col: 17, offset: 1772},
							val:        "[+-]",
							chars:      []rune{'+', '-'},
							ignoreCase: false,
							inverted:   false,
						},
					},
					&oneOrMoreExpr{
						pos: position{line: 70, col: 23, offset: 1778},
						expr: &ruleRefExpr{
							pos:  position{line: 70, col: 23, offset: 1778},
							name: "DecimalDigit",
						},
					},
				},
			},
		},
		{
			name: "String",
			pos:  position{line: 72, col: 1, offset: 1795},
			expr: &actionExpr{
				pos: position{line: 72, col: 10, offset: 1806},
				run: (*parser).callonString1,
				expr: &seqExpr{
					pos: position{line: 72, col: 10, offset: 1806},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 72, col: 10, offset: 1806},
							val:        "\"",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 72, col: 14, offset: 1810},
							expr: &choiceExpr{
								pos: position{line: 72, col: 16, offset: 1812},
								alternatives: []interface{}{
									&seqExpr{
										pos: position{line: 72, col: 16, offset: 1812},
										exprs: []interface{}{
											&notExpr{
												pos: position{line: 72, col: 16, offset: 1812},
												expr: &ruleRefExpr{
													pos:  position{line: 72, col: 17, offset: 1813},
													name: "EscapedChar",
												},
											},
											&anyMatcher{
												line: 72, col: 29, offset: 1825,
											},
										},
									},
									&seqExpr{
										pos: position{line: 72, col: 33, offset: 1829},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 72, col: 33, offset: 1829},
												val:        "\\",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 72, col: 38, offset: 1834},
												name: "EscapeSequence",
											},
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 72, col: 56, offset: 1852},
							val:        "\"",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "EscapedChar",
			pos:  position{line: 78, col: 1, offset: 2030},
			expr: &charClassMatcher{
				pos:        position{line: 78, col: 15, offset: 2046},
				val:        "[\\x00-\\x1f\"\\\\]",
				chars:      []rune{'"', '\\'},
				ranges:     []rune{'\x00', '\x1f'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EscapeSequence",
			pos:  position{line: 80, col: 1, offset: 2064},
			expr: &choiceExpr{
				pos: position{line: 80, col: 18, offset: 2083},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 80, col: 18, offset: 2083},
						name: "SingleCharEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 80, col: 37, offset: 2102},
						name: "UnicodeEscape",
					},
				},
			},
		},
		{
			name: "SingleCharEscape",
			pos:  position{line: 82, col: 1, offset: 2119},
			expr: &charClassMatcher{
				pos:        position{line: 82, col: 20, offset: 2140},
				val:        "[\"\\\\/bfnrt]",
				chars:      []rune{'"', '\\', '/', 'b', 'f', 'n', 'r', 't'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "UnicodeEscape",
			pos:  position{line: 84, col: 1, offset: 2155},
			expr: &seqExpr{
				pos: position{line: 84, col: 17, offset: 2173},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 84, col: 17, offset: 2173},
						val:        "u",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 84, col: 21, offset: 2177},
						name: "HexDigit",
					},
					&ruleRefExpr{
						pos:  position{line: 84, col: 30, offset: 2186},
						name: "HexDigit",
					},
					&ruleRefExpr{
						pos:  position{line: 84, col: 39, offset: 2195},
						name: "HexDigit",
					},
					&ruleRefExpr{
						pos:  position{line: 84, col: 48, offset: 2204},
						name: "HexDigit",
					},
				},
			},
		},
		{
			name: "DecimalDigit",
			pos:  position{line: 86, col: 1, offset: 2216},
			expr: &charClassMatcher{
				pos:        position{line: 86, col: 16, offset: 2233},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "NonZeroDecimalDigit",
			pos:  position{line: 88, col: 1, offset: 2242},
			expr: &charClassMatcher{
				pos:        position{line: 88, col: 23, offset: 2266},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "HexDigit",
			pos:  position{line: 90, col: 1, offset: 2275},
			expr: &charClassMatcher{
				pos:        position{line: 90, col: 12, offset: 2288},
				val:        "[0-9a-f]i",
				ranges:     []rune{'0', '9', 'a', 'f'},
				ignoreCase: true,
				inverted:   false,
			},
		},
		{
			name: "Bool",
			pos:  position{line: 92, col: 1, offset: 2301},
			expr: &choiceExpr{
				pos: position{line: 92, col: 8, offset: 2310},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 92, col: 8, offset: 2310},
						run: (*parser).callonBool2,
						expr: &litMatcher{
							pos:        position{line: 92, col: 8, offset: 2310},
							val:        "true",
							ignoreCase: false,
						},
					},
					&actionExpr{
						pos: position{line: 92, col: 38, offset: 2340},
						run: (*parser).callonBool4,
						expr: &litMatcher{
							pos:        position{line: 92, col: 38, offset: 2340},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Null",
			pos:  position{line: 94, col: 1, offset: 2373},
			expr: &actionExpr{
				pos: position{line: 94, col: 8, offset: 2382},
				run: (*parser).callonNull1,
				expr: &litMatcher{
					pos:        position{line: 94, col: 8, offset: 2382},
					val:        "null",
					ignoreCase: false,
				},
			},
		},
		{
			name:        "_",
			displayName: "\"whitespace\"",
			pos:         position{line: 96, col: 1, offset: 2412},
			expr: &zeroOrMoreExpr{
				pos: position{line: 96, col: 18, offset: 2431},
				expr: &charClassMatcher{
					pos:        position{line: 96, col: 18, offset: 2431},
					val:        "[ \\t\\r\\n]",
					chars:      []rune{' ', '\t', '\r', '\n'},
					ignoreCase: false,
					inverted:   false,
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 98, col: 1, offset: 2445},
			expr: &notExpr{
				pos: position{line: 98, col: 7, offset: 2453},
				expr: &anyMatcher{
					line: 98, col: 8, offset: 2454,
				},
			},
		},
	},
}

func (c *current) onJSON1(vals interface{}) (interface{}, error) {

	valsSl := toIfaceSlice(vals)
	switch len(valsSl) {
	case 0:
		return nil, nil
	case 1:
		return valsSl[0], nil
	default:
		return valsSl, nil
	}
}

func (p *parser) callonJSON1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onJSON1(stack["vals"])
}

func (c *current) onValue1(val interface{}) (interface{}, error) {

	return val, nil
}

func (p *parser) callonValue1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onValue1(stack["val"])
}

func (c *current) onObject1(vals interface{}) (interface{}, error) {

	res := make(map[string]interface{})
	valsSl := toIfaceSlice(vals)
	if len(valsSl) == 0 {
		return res, nil
	}
	res[valsSl[0].(string)] = valsSl[4]
	restSl := toIfaceSlice(valsSl[5])
	for _, v := range restSl {
		vSl := toIfaceSlice(v)
		res[vSl[2].(string)] = vSl[6]
	}
	return res, nil
}

func (p *parser) callonObject1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onObject1(stack["vals"])
}

func (c *current) onArray1(vals interface{}) (interface{}, error) {

	valsSl := toIfaceSlice(vals)
	if len(valsSl) == 0 {
		return []interface{}{}, nil
	}
	res := []interface{}{valsSl[0]}
	restSl := toIfaceSlice(valsSl[1])
	for _, v := range restSl {
		vSl := toIfaceSlice(v)
		res = append(res, vSl[2])
	}
	return res, nil
}

func (p *parser) callonArray1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onArray1(stack["vals"])
}

func (c *current) onNumber1() (interface{}, error) {

	// JSON numbers have the same syntax as Go's, and are parseable using
	// strconv.
	return strconv.ParseFloat(string(c.text), 64)
}

func (p *parser) callonNumber1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNumber1()
}

func (c *current) onString1() (interface{}, error) {

	// TODO : the forward slash (solidus) is not a valid escape in Go, it will
	// fail if there's one in the string
	return strconv.Unquote(string(c.text))
}

func (p *parser) callonString1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onString1()
}

func (c *current) onBool2() (interface{}, error) {
	return true, nil
}

func (p *parser) callonBool2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBool2()
}

func (c *current) onBool4() (interface{}, error) {
	return false, nil
}

func (p *parser) callonBool4() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBool4()
}

func (c *current) onNull1() (interface{}, error) {
	return nil, nil
}

func (p *parser) callonNull1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNull1()
}

var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule = errors.New("grammar has no rule")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// Debug creates an Option to set the debug flag to b. When set to true,
// debugging information is printed to stdout while parsing.
//
// The default is false.
func Debug(b bool) Option {
	return func(p *parser) Option {
		old := p.debug
		p.debug = b
		return Debug(old)
	}
}

// Memoize creates an Option to set the memoize flag to b. When set to true,
// the parser will cache all results so each expression is evaluated only
// once. This guarantees linear parsing time even for pathological cases,
// at the expense of more memory and slower times for typical cases.
//
// The default is false.
func Memoize(b bool) Option {
	return func(p *parser) Option {
		old := p.memoize
		p.memoize = b
		return Memoize(old)
	}
}

// Recover creates an Option to set the recover flag to b. When set to
// true, this causes the parser to recover from panics and convert it
// to an error. Setting it to false can be useful while debugging to
// access the full stack trace.
//
// The default is true.
func Recover(b bool) Option {
	return func(p *parser) Option {
		old := p.recover
		p.recover = b
		return Recover(old)
	}
}

// GlobalStore creates an Option to set a key to a certain value in
// the globalStore.
func GlobalStore(key string, value interface{}) Option {
	return func(p *parser) Option {
		old := p.cur.globalStore[key]
		p.cur.globalStore[key] = value
		return GlobalStore(key, old)
	}
}

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (i interface{}, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = closeErr
		}
	}()
	return ParseReader(filename, f, opts...)
}

// ParseReader parses the data from r using filename as information in the
// error messages.
func ParseReader(filename string, r io.Reader, opts ...Option) (interface{}, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Parse(filename, b, opts...)
}

// Parse parses the data from b using filename as information in the
// error messages.
func Parse(filename string, b []byte, opts ...Option) (interface{}, error) {
	return newParser(filename, b, opts...).parse(g)
}

// position records a position in the text.
type position struct {
	line, col, offset int
}

func (p position) String() string {
	return fmt.Sprintf("%d:%d [%d]", p.line, p.col, p.offset)
}

// savepoint stores all state required to go back to this point in the
// parser.
type savepoint struct {
	position
	rn rune
	w  int
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match

	// the state allows the parser to store arbitrary values and rollback them if needed
	state statedict
	// the globalStore allows the parser to store arbitrary values
	globalStore map[string]interface{}
}

type statedict map[string]interface{}

// the AST types...

type grammar struct {
	pos   position
	rules []*rule
}

type rule struct {
	pos         position
	name        string
	displayName string
	expr        interface{}
}

type choiceExpr struct {
	pos          position
	alternatives []interface{}
}

type actionExpr struct {
	pos  position
	expr interface{}
	run  func(*parser) (interface{}, error)
}

type seqExpr struct {
	pos   position
	exprs []interface{}
}

type labeledExpr struct {
	pos   position
	label string
	expr  interface{}
}

type expr struct {
	pos  position
	expr interface{}
}

type andExpr expr
type notExpr expr
type zeroOrOneExpr expr
type zeroOrMoreExpr expr
type oneOrMoreExpr expr

type ruleRefExpr struct {
	pos  position
	name string
}

type andCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type notCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type litMatcher struct {
	pos        position
	val        string
	ignoreCase bool
}

type charClassMatcher struct {
	pos             position
	val             string
	basicLatinChars [128]bool
	chars           []rune
	ranges          []rune
	classes         []*unicode.RangeTable
	ignoreCase      bool
	inverted        bool
}

type anyMatcher position

// errList cumulates the errors found by the parser.
type errList []error

func (e *errList) add(err error) {
	*e = append(*e, err)
}

func (e errList) err() error {
	if len(e) == 0 {
		return nil
	}
	e.dedupe()
	return e
}

func (e *errList) dedupe() {
	var cleaned []error
	set := make(map[string]bool)
	for _, err := range *e {
		if msg := err.Error(); !set[msg] {
			set[msg] = true
			cleaned = append(cleaned, err)
		}
	}
	*e = cleaned
}

func (e errList) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		var buf bytes.Buffer

		for i, err := range e {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}
}

// parserError wraps an error with a prefix indicating the rule in which
// the error occurred. The original error is stored in the Inner field.
type parserError struct {
	Inner    error
	pos      position
	prefix   string
	expected []string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	p := &parser{
		filename: filename,
		errs:     new(errList),
		data:     b,
		pt:       savepoint{position: position{line: 1}},
		recover:  true,
		cur: current{
			state:       make(statedict),
			globalStore: make(map[string]interface{}),
		},
		maxFailPos:      position{col: 1, line: 1},
		maxFailExpected: make([]string, 0, 20),
	}
	p.setOptions(opts)
	return p
}

// setOptions applies the options to the parser.
func (p *parser) setOptions(opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}

type resultTuple struct {
	v   interface{}
	b   bool
	end savepoint
}

type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	depth   int
	recover bool
	debug   bool

	memoize bool
	// memoization table for the packrat algorithm:
	// map[offset in source] map[expression or rule] {value, match}
	memo map[int]map[interface{}]resultTuple

	// rules table, maps the rule identifier to the rule node
	rules map[string]*rule
	// variables stack, map of label to value
	vstack []map[string]interface{}
	// rule stack, allows identification of the current rule in errors
	rstack []*rule

	// stats
	exprCnt int

	// parse fail
	maxFailPos            position
	maxFailExpected       []string
	maxFailInvertExpected bool
}

// push a variable set on the vstack.
func (p *parser) pushV() {
	if cap(p.vstack) == len(p.vstack) {
		// create new empty slot in the stack
		p.vstack = append(p.vstack, nil)
	} else {
		// slice to 1 more
		p.vstack = p.vstack[:len(p.vstack)+1]
	}

	// get the last args set
	m := p.vstack[len(p.vstack)-1]
	if m != nil && len(m) == 0 {
		// empty map, all good
		return
	}

	m = make(map[string]interface{})
	p.vstack[len(p.vstack)-1] = m
}

// pop a variable set from the vstack.
func (p *parser) popV() {
	// if the map is not empty, clear it
	m := p.vstack[len(p.vstack)-1]
	if len(m) > 0 {
		// GC that map
		p.vstack[len(p.vstack)-1] = nil
	}
	p.vstack = p.vstack[:len(p.vstack)-1]
}

func (p *parser) print(prefix, s string) string {
	if !p.debug {
		return s
	}

	fmt.Printf("%s %d:%d:%d: %s [%#U]\n",
		prefix, p.pt.line, p.pt.col, p.pt.offset, s, p.pt.rn)
	return s
}

func (p *parser) in(s string) string {
	p.depth++
	return p.print(strings.Repeat(" ", p.depth)+">", s)
}

func (p *parser) out(s string) string {
	p.depth--
	return p.print(strings.Repeat(" ", p.depth)+"<", s)
}

func (p *parser) addErr(err error) {
	p.addErrAt(err, p.pt.position, []string{})
}

func (p *parser) addErrAt(err error, pos position, expected []string) {
	var buf bytes.Buffer
	if p.filename != "" {
		buf.WriteString(p.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%d:%d (%d)", pos.line, pos.col, pos.offset))
	if len(p.rstack) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(": ")
		}
		rule := p.rstack[len(p.rstack)-1]
		if rule.displayName != "" {
			buf.WriteString("rule " + rule.displayName)
		} else {
			buf.WriteString("rule " + rule.name)
		}
	}
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String(), expected: expected}
	p.errs.add(pe)
}

func (p *parser) failAt(fail bool, pos position, want string) {
	// process fail if parsing fails and not inverted or parsing succeeds and invert is set
	if fail == p.maxFailInvertExpected {
		if pos.offset < p.maxFailPos.offset {
			return
		}

		if pos.offset > p.maxFailPos.offset {
			p.maxFailPos = pos
			p.maxFailExpected = p.maxFailExpected[:0]
		}

		if p.maxFailInvertExpected {
			want = "!" + want
		}
		p.maxFailExpected = append(p.maxFailExpected, want)
	}
}

// read advances the parser to the next rune.
func (p *parser) read() {
	p.pt.offset += p.pt.w
	rn, n := utf8.DecodeRune(p.data[p.pt.offset:])
	p.pt.rn = rn
	p.pt.w = n
	p.pt.col++
	if rn == '\n' {
		p.pt.line++
		p.pt.col = 0
	}

	if rn == utf8.RuneError {
		if n == 1 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// restore parser position to the savepoint pt.
func (p *parser) restore(pt savepoint) {
	if p.debug {
		defer p.out(p.in("restore"))
	}
	if pt.offset == p.pt.offset {
		return
	}
	p.pt = pt
}

// copy and return parser current state.
func (p *parser) copyState() (state statedict) {
	if p.debug {
		defer p.out(p.in("copyState"))
	}
	state = make(statedict)
	for k, v := range p.cur.state {
		state[k] = v
	}
	return state
}

// restore parser current state to the state statedict.
func (p *parser) restoreState(state statedict) {
	if p.debug {
		defer p.out(p.in("restoreState"))
	}
	p.cur.state = state
}

// get the slice of bytes from the savepoint start to the current position.
func (p *parser) sliceFrom(start savepoint) []byte {
	return p.data[start.position.offset:p.pt.position.offset]
}

func (p *parser) getMemoized(node interface{}) (resultTuple, bool) {
	if len(p.memo) == 0 {
		return resultTuple{}, false
	}
	m := p.memo[p.pt.offset]
	if len(m) == 0 {
		return resultTuple{}, false
	}
	res, ok := m[node]
	return res, ok
}

func (p *parser) setMemoized(pt savepoint, node interface{}, tuple resultTuple) {
	if p.memo == nil {
		p.memo = make(map[int]map[interface{}]resultTuple)
	}
	m := p.memo[pt.offset]
	if m == nil {
		m = make(map[interface{}]resultTuple)
		p.memo[pt.offset] = m
	}
	m[node] = tuple
}

func (p *parser) buildRulesTable(g *grammar) {
	p.rules = make(map[string]*rule, len(g.rules))
	for _, r := range g.rules {
		p.rules[r.name] = r
	}
}

func (p *parser) parse(g *grammar) (val interface{}, err error) {
	if len(g.rules) == 0 {
		p.addErr(errNoRule)
		return nil, p.errs.err()
	}

	// TODO : not super critical but this could be generated
	p.buildRulesTable(g)

	if p.recover {
		// panic can be used in action code to stop parsing immediately
		// and return the panic as an error.
		defer func() {
			if e := recover(); e != nil {
				if p.debug {
					defer p.out(p.in("panic handler"))
				}
				val = nil
				switch e := e.(type) {
				case error:
					p.addErr(e)
				default:
					p.addErr(fmt.Errorf("%v", e))
				}
				err = p.errs.err()
			}
		}()
	}

	// start rule is rule [0]
	p.read() // advance to first rune
	val, ok := p.parseRule(g.rules[0])
	if !ok {
		if len(*p.errs) == 0 {
			// If parsing fails, but no errors have been recorded, the expected values
			// for the farthest parser position are returned as error.
			maxFailExpectedMap := make(map[string]struct{}, len(p.maxFailExpected))
			for _, v := range p.maxFailExpected {
				maxFailExpectedMap[v] = struct{}{}
			}
			expected := make([]string, 0, len(maxFailExpectedMap))
			eof := false
			if _, ok := maxFailExpectedMap["!."]; ok {
				delete(maxFailExpectedMap, "!.")
				eof = true
			}
			for k := range maxFailExpectedMap {
				expected = append(expected, k)
			}
			sort.Strings(expected)
			if eof {
				expected = append(expected, "EOF")
			}
			p.addErrAt(errors.New("no match found, expected: "+listJoin(expected, ", ", "or")), p.maxFailPos, expected)
		}
		return nil, p.errs.err()
	}
	return val, p.errs.err()
}

func listJoin(list []string, sep string, lastSep string) string {
	switch len(list) {
	case 0:
		return ""
	case 1:
		return list[0]
	default:
		return fmt.Sprintf("%s %s %s", strings.Join(list[:len(list)-1], sep), lastSep, list[len(list)-1])
	}
}

func (p *parser) parseRule(rule *rule) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRule " + rule.name))
	}

	if p.memoize {
		res, ok := p.getMemoized(rule)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
	}

	start := p.pt
	p.rstack = append(p.rstack, rule)
	p.pushV()
	val, ok := p.parseExpr(rule.expr)
	p.popV()
	p.rstack = p.rstack[:len(p.rstack)-1]
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}

	if p.memoize {
		p.setMemoized(start, rule, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseExpr(expr interface{}) (interface{}, bool) {
	var pt savepoint

	if p.memoize {
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	p.exprCnt++
	var val interface{}
	var ok bool
	switch expr := expr.(type) {
	case *actionExpr:
		val, ok = p.parseActionExpr(expr)
	case *andCodeExpr:
		val, ok = p.parseAndCodeExpr(expr)
	case *andExpr:
		val, ok = p.parseAndExpr(expr)
	case *anyMatcher:
		val, ok = p.parseAnyMatcher(expr)
	case *charClassMatcher:
		val, ok = p.parseCharClassMatcher(expr)
	case *choiceExpr:
		val, ok = p.parseChoiceExpr(expr)
	case *labeledExpr:
		val, ok = p.parseLabeledExpr(expr)
	case *litMatcher:
		val, ok = p.parseLitMatcher(expr)
	case *notCodeExpr:
		val, ok = p.parseNotCodeExpr(expr)
	case *notExpr:
		val, ok = p.parseNotExpr(expr)
	case *oneOrMoreExpr:
		val, ok = p.parseOneOrMoreExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *zeroOrMoreExpr:
		val, ok = p.parseZeroOrMoreExpr(expr)
	case *zeroOrOneExpr:
		val, ok = p.parseZeroOrOneExpr(expr)
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
	if p.memoize {
		p.setMemoized(pt, expr, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseActionExpr(act *actionExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseActionExpr"))
	}

	start := p.pt
	val, ok := p.parseExpr(act.expr)
	if ok {
		p.cur.pos = start.position
		p.cur.text = p.sliceFrom(start)
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position, []string{})
		}
		val = actVal
	}
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}
	return val, ok
}

func (p *parser) parseAndCodeExpr(and *andCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndCodeExpr"))
	}

	ok, err := and.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, ok
}

func (p *parser) parseAndExpr(and *andExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	pt := p.pt
	state := p.copyState()
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restore(pt)
	p.restoreState(state)
	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	if p.pt.rn != utf8.RuneError {
		start := p.pt
		p.read()
		p.failAt(true, start.position, ".")
		return p.sliceFrom(start), true
	}
	p.failAt(false, p.pt.position, ".")
	return nil, false
}

func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	cur := p.pt.rn
	start := p.pt

	// can't match EOF
	if cur == utf8.RuneError {
		p.failAt(false, start.position, chr.val)
		return nil, false
	}

	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		p.failAt(true, start.position, chr.val)
		return p.sliceFrom(start), true
	}
	p.failAt(false, start.position, chr.val)
	return nil, false
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for _, alt := range ch.alternatives {
		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			return val, ok
		}
	}
	return nil, false
}

func (p *parser) parseLabeledExpr(lab *labeledExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLabeledExpr"))
	}

	p.pushV()
	val, ok := p.parseExpr(lab.expr)
	p.popV()
	if ok && lab.label != "" {
		m := p.vstack[len(p.vstack)-1]
		m[lab.label] = val
	}
	return val, ok
}

func (p *parser) parseLitMatcher(lit *litMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLitMatcher"))
	}

	ignoreCase := ""
	if lit.ignoreCase {
		ignoreCase = "i"
	}
	val := fmt.Sprintf("%q%s", lit.val, ignoreCase)
	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.failAt(false, start.position, val)
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	p.failAt(true, start.position, val)
	return p.sliceFrom(start), true
}

func (p *parser) parseNotCodeExpr(not *notCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotCodeExpr"))
	}

	ok, err := not.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, !ok
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	state := p.copyState()
	p.pushV()
	p.maxFailInvertExpected = !p.maxFailInvertExpected
	_, ok := p.parseExpr(not.expr)
	p.maxFailInvertExpected = !p.maxFailInvertExpected
	p.popV()
	p.restore(pt)
	p.restoreState(state)
	return nil, !ok
}

func (p *parser) parseOneOrMoreExpr(expr *oneOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseOneOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			if len(vals) == 0 {
				// did not match once, no match
				return nil, false
			}
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseRuleRefExpr(ref *ruleRefExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRuleRefExpr " + ref.name))
	}

	if ref.name == "" {
		panic(fmt.Sprintf("%s: invalid rule: missing name", ref.pos))
	}

	rule := p.rules[ref.name]
	if rule == nil {
		p.addErr(fmt.Errorf("undefined rule: %s", ref.name))
		return nil, false
	}
	return p.parseRule(rule)
}

func (p *parser) parseSeqExpr(seq *seqExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseSeqExpr"))
	}

	vals := make([]interface{}, 0, len(seq.exprs))

	pt := p.pt
	state := p.copyState()
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restore(pt)
			p.restoreState(state)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
}

func (p *parser) parseZeroOrMoreExpr(expr *zeroOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseZeroOrOneExpr(expr *zeroOrOneExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrOneExpr"))
	}

	p.pushV()
	val, _ := p.parseExpr(expr.expr)
	p.popV()
	// whether it matched or not, consider it a match
	return val, true
}
