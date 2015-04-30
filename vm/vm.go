package vm

import (
	"bytes"
	"fmt"
)

// ϡtheProgram is the variable that holds the program generated by the
// builder for the input PEG.
var ϡtheProgram *ϡprogram

//+ϡ following code is part of the generated parser

// ϡsentinel is a type used to define sentinel values that shouldn't
// be equal to something else.
type ϡsentinel int

const (
	// ϡmatchFailed is a sentinel value used to indicate a match failure.
	ϡmatchFailed ϡsentinel = iota
)

const (
	// stack IDs, used in PUSH and POP's first argument
	ϡpstackID = iota + 1
	ϡlstackID
	ϡvstackID
	ϡistackID

	ϡvValNil    = 0
	ϡvValFailed = 1
	ϡvValEmpty  = 2
)

// special values that may be pushed on the V stack.
var ϡvSpecialValues = []interface{}{
	nil,
	ϡmatchFailed,
	[]interface{}(nil),
}

type ϡmemoizedResult struct {
	v  interface{}
	pt ϡsvpt
}

// ϡprogram is the data structure that is generated by the builder
// based on an input PEG. It contains the program information required
// to execute the grammar using the vm.
type ϡprogram struct {
	instrs []ϡinstr

	// lists
	ms []ϡmatcher
	as []func(*ϡvm) (interface{}, error)
	bs []func(*ϡvm) (bool, error)
	ss []string

	// instrToRule is the mapping of an instruction index to a rule
	// identifier (or display name) in the ss list:
	//
	// ss[instrToRule[instrIndex]] == name of the rule
	//
	// Since instructions are limited to 65535, the size of this slice
	// is bounded.
	instrToRule []int
}

// ruleNameAt returns the name of the rule that contains the instruction
// index. It returns an empty string is the instruction is not part of a
// rule (bootstrap instruction, invalid index).
func (pg ϡprogram) ruleNameAt(instrIx int) string {
	if instrIx < 0 || instrIx >= len(pg.instrToRule) {
		return ""
	}
	ssIx := pg.instrToRule[instrIx]
	if ssIx < 0 || ssIx >= len(pg.ss) {
		return ""
	}
	return pg.ss[ssIx]
}

type ϡvm struct {
	// input
	filename string
	parser   *ϡparser

	// options
	debug   bool
	memoize bool
	recover bool
	// TODO : no bounds checking option (for stacks)? benchmark to see if it's worth it.

	// program data
	pc    int
	depth int
	pg    *ϡprogram
	cur   current

	// stacks
	p       ϡpstack
	l       ϡlstack
	v       ϡvstack
	i       ϡistack
	varSets []map[string]interface{}

	// TODO: memoization...
	// TODO: farthest failure position

	// error list
	errs errList
}

// setOptions applies the options in sequence on the vm. It returns the
// vm to allow for chaining calls.
func (v *ϡvm) setOptions(opts []Option) *ϡvm {
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// addErr adds the error at the current parser position, without rule name
// information.
func (v *ϡvm) addErr(err error) {
	v.addErrAt(err, -1, v.parser.pt.position)
}

// addErrAt adds the error at the specified position, for the instruction
// at instrIx.
func (v *ϡvm) addErrAt(err error, instrIx int, pos position) {
	var buf bytes.Buffer
	if v.filename != "" {
		buf.WriteString(v.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%s", pos))

	ruleNm := v.pg.ruleNameAt(instrIx)
	if ruleNm != "" {
		buf.WriteString(": ")
		buf.WriteString("rule " + ruleNm)
	}

	pe := &parserError{Inner: err, ϡprefix: buf.String()}
	v.errs.ϡadd(pe)
}

// injectDebug injects debugging opcodes into the list of instructions.
func (v *ϡvm) injectDebug() {
	var op ϡop
	var n int

	instrs, _ := ϡencodeInstr(ϡopDebug)
	dbgInstr := instrs[0]

	for i := 0; i < len(v.pg.instrs); i++ {
		if n > 0 {
			n -= 4
			continue
		}
		op, n, _, _, _ = v.pg.instrs[i].decode()
		n -= 3
		switch op {
		case ϡopCall, ϡopCallA, ϡopCallB:
			v.pg.instrs = append(v.pg.instrs[:i], append([]ϡinstr{dbgInstr},
				v.pg.instrs[i:]...)...)
			i++
		}
	}
}

// run executes the provided program in this VM, and returns the result.
func (v *ϡvm) run(pg *ϡprogram) (interface{}, error) {
	v.pg = pg
	if v.debug {
		// inject debug opcodes before each CALL{A,B}
		v.injectDebug()
	}
	ret := v.dispatch()

	// if the match failed, translate that to a nil result and make
	// sure it returns an error
	if ret == ϡmatchFailed {
		ret = nil
		if len(v.errs) == 0 {
			v.addErr(errNoMatch)
		}
	}

	return ret, v.errs.ϡerr()
}

// dispatch is the proper execution method of the VM, it loops over
// the instructions and executes each opcode.
func (v *ϡvm) dispatch() interface{} {
	for {
		// fetch and decode the instruction
		instr := v.pg.instrs[v.pc]
		op, n, a0, a1, a2 := instr.decode()

		// increment program counter
		v.pc++

		switch op {
		case ϡopCall:
			ix := v.i.pop()
			v.i.push(v.pc)
			v.pc = ix
			v.depth++

		case ϡopCallA:
			v.v.pop()
			start := v.p.pop()
			v.cur.pos = start.position
			v.cur.text = v.parser.sliceFrom(start)
			if a0 >= len(v.pg.as) {
				panic(fmt.Sprintf("invalid %s argument: %d", op, a0))
			}
			fn := v.pg.as[a0]
			val, err := fn(v)
			if err != nil {
				v.addErrAt(err, v.pc-1, start.position)
			}
			v.v.push(val)

		case ϡopCallB:
			v.cur.pos = v.parser.pt.position
			v.cur.text = nil
			if a0 >= len(v.pg.bs) {
				panic(fmt.Sprintf("invalid %s argument: %d", op, a0))
			}
			fn := v.pg.bs[a0]
			val, err := fn(v)
			if err != nil {
				v.addErrAt(err, v.pc-1, v.parser.pt.position)
			}
			if !val {
				v.v.push(ϡmatchFailed)
				break
			}
			v.v.push(nil)

		case ϡopCumulOrF:
			va, vb := v.v.pop(), v.v.pop()
			if va == ϡmatchFailed {
				v.v.push(ϡmatchFailed)
				break
			}
			switch vb := vb.(type) {
			case []interface{}:
				vb = append(vb, va)
				v.v.push(vb)
			case ϡsentinel:
				v.v.push([]interface{}{va})
			default:
				panic(fmt.Sprintf("invalid %s value type on the V stack: %T", op, vb))
			}

		case ϡopDebug:
			// TODO : print n instructions above and below, stacks, decode args
			fmt.Println("hello world")

		case ϡopExit:
			return v.v.pop()

		case ϡopNilIfF:
			if top := v.v.pop(); top == ϡmatchFailed {
				v.v.push(nil)
				break
			}
			v.v.push(ϡmatchFailed)

		case ϡopNilIfT:
			if top := v.v.pop(); top != ϡmatchFailed {
				v.v.push(nil)
				break
			}
			v.v.push(ϡmatchFailed)

		case ϡopJump:
			v.pc = a0

		case ϡopJumpIfF:
			if top := v.v.peek(); top == ϡmatchFailed {
				v.pc = a0
			}

		case ϡopJumpIfT:
			if top := v.v.peek(); top != ϡmatchFailed {
				v.pc = a0
			}

		case ϡopMatch:
			start := v.parser.pt
			if a0 >= len(v.pg.ms) {
				panic(fmt.Sprintf("invalid %s argument: %d", op, a0))
			}
			m := v.pg.ms[a0]
			if ok := m.match(v.parser); ok {
				v.v.push(v.parser.sliceFrom(start))
				break
			}
			v.v.push(ϡmatchFailed)
			v.parser.pt = start

		case ϡopPop:
			switch a0 {
			case ϡlstackID:
				v.l.pop()
			case ϡpstackID:
				v.p.pop()
			default:
				panic(fmt.Sprintf("invalid %s argument: %d", op, a0))
			}

		case ϡopPopVJumpIfF:
			if top := v.v.peek(); top == ϡmatchFailed {
				v.v.pop()
				v.pc = a0
			}

		case ϡopPush:
			switch a0 {
			case ϡpstackID:
				v.p.push(v.parser.pt)
			case ϡistackID:
				v.i.push(a1)
			case ϡvstackID:
				if a1 >= len(ϡvSpecialValues) {
					panic(fmt.Sprintf("invalid %s V stack argument: %d", op, a1))
				}
				v.v.push(ϡvSpecialValues[a1])
			case ϡlstackID:
				ar := make([]int, n)
				src := []int{0, 0, a2, a1}
				for i := 0; i < n; i++ {
					lsrc := len(src)
					ar[i] = src[lsrc-1]
					src = src[:lsrc-1]
					if lsrc-1 == 0 && i < n-1 {
						// need more
						instr := v.pg.instrs[v.pc]
						a0, a1, a2, a3 := instr.decodeLs()
						src = append(src, a3, a2, a1, a0)
						v.pc++
					}
				}
				v.l.push(ar)
			default:
				panic(fmt.Sprintf("invalid %s argument: %d", op, a0))
			}

		case ϡopRestore:
			pt := v.p.pop()
			v.parser.pt = pt

		case ϡopRestoreIfF:
			pt := v.p.pop()
			if top := v.v.peek(); top == ϡmatchFailed {
				v.parser.pt = pt
			}

		case ϡopReturn:
			ix := v.i.pop()
			v.pc = ix

			// clean-up the varSet, if required
			if v.depth == len(v.varSets)-1 {
				if m := v.varSets[v.depth]; len(m) > 0 {
					v.varSets[v.depth] = nil
				}
			}

			v.depth--
			if v.depth < 0 {
				panic("negative call depth")
			}

		case ϡopStoreIfT:
			if top := v.v.peek(); top != ϡmatchFailed {
				// make sure the var set for this depth level is available
				// TODO : this is not correct, depth based var stack doesn't work,
				// do similar to gen code.
				if v.depth >= len(v.varSets) {
					// grow varSets to v.depth+1
					if v.depth < cap(v.varSets) {
						v.varSets = v.varSets[:v.depth+1]
						v.varSets = v.varSets[:v.depth+1]
					} else {
						newSets := make([]map[string]interface{}, v.depth+1)
						copy(newSets, v.varSets)
						v.varSets = newSets
					}
				}

				// create the var set map if required
				m := v.varSets[v.depth]
				if m == nil {
					m = make(map[string]interface{})
					v.varSets[v.depth] = m
				}

				// get the label name
				if a0 >= len(v.pg.ss) {
					panic(fmt.Sprintf("invalid %s argument: %d", op, a0))
				}
				lbl := v.pg.ss[a0]

				// store the value
				m[lbl] = top
			}

		case ϡopTakeLOrJump:
			ix := v.l.take()
			if ix < 0 {
				v.pc = a0
				break
			}
			v.i.push(ix)

		default:
			panic(fmt.Sprintf("unknown opcode %s", op))
		}
	}
}
