package parser

import (
	"encoding/json"
	"reflect"
	"runtime"
	"strings"
)

func ScanString(s string, start int) string {
	quotationMark := s[start : start+1]
	var loop func(int, bool, string) string
	loop = func(start int, skip bool, str string) string {
		curChar := s[start : start+1]
		if start == len(s) {
			return ""
		} else if skip {
			return loop(1+start, false, str+curChar)
		} else if curChar == "\\" {
			return loop(1+start, true, str+"\\")
		} else if curChar == quotationMark {
			return str
		}
		return loop(1+start, false, str+curChar)
	}
	return loop(1+start, false, "")
}

func Scan(s string) []*Node {
	var scan1 func(string, int) (*Node, int)
	scan1 = func(s string, start int) (*Node, int) {
		if start == len(s) {
			return &Node{Type: EofType}, start
		}
		if StartWithOneOf(s, start, significant_whitespaces) != "" {
			return NewNode(NewlineType, start, start+1, nil, "", 0, nil), start + 1
		}
		if IsWhitespace(s[start : start+1]) {
			return scan1(s, 1+start)
		}
		if StartWithOneOf(s, start, line_comment) != "" {
			lineEnd := FindNext(s, start, func(s string, start int) bool {
				return s[start:start+1] == "\n"
			})
			return NewNode(CommentType, start, 1+lineEnd, nil, s[start:lineEnd], 0, nil), lineEnd
		}
		if StartWith(s, start, comment_start) != "" {
			lineEnd := FindNext(s, start, func(s string, start int) bool {
				return StartWith(s, start, comment_end) != ""
			})
			end := lineEnd + len(comment_end)
			return NewNode(CommentType, start, end, nil, s[start:end], 0, nil), end
		}
		if delim := FindDelim(s, start); delim != "" {
			end := start + len(delim)
			return NewNode(TokenType, start, end, nil, delim, 0, nil), end
		}
		if op := FindOperator(s, start); op != "" {
			end := start + len(op)
			return NewNode(TokenType, start, end, nil, op, 0, nil), end
		}
		if StartWithOneOf(s, start, quotation_marks) != "" {
			str := ScanString(s, start)
			end := start + len(str) + 2
			return NewNode(StrType, start, end, nil, str, 0, nil), end
		}
		if StartWithOneOf(s, start, lisp_char) != "" {
			if len(s) <= 2+start {
				panic("scan-string: reached EOF while scanning char")
			}
			var loop func(int) int
			loop = func(end int) int {
				if IsWhitespace(s[end:end+1]) || IsDelim(s[end:end+1]) {
					return end
				} else {
					return loop(end + 1)
				}
			}
			end := loop(3 + start)
			return NewNode(CharacterType, start, end, nil, s[end-1:end], 0, nil), end
		}
		var loop func(int, string) (*Node, int)
		loop = func(pos int, chars string) (*Node, int) {
			if len(s) <= pos ||
				IsWhitespace(s[pos:pos+1]) ||
				FindDelim(s, pos) != "" ||
				FindOperator(s, pos) != "" {
				return NewNode(TokenType, start, pos, nil, chars, 0, nil), pos
			} else {
				return loop(1+pos, chars+s[pos:pos+1])
			}
		}
		return loop(start, "")
	}
	var loop func(int, []*Node) []*Node
	loop = func(start int, toks []*Node) []*Node {
		tok, newStart := scan1(s, start)
		if tok.Type == EofType {
			return toks
		} else {
			return loop(newStart, append(toks, tok))
		}
	}
	return loop(0, make([]*Node, 0))
}

type Pair struct {
	combinator Combinator
	toks       []*Node
}

func filter(f func(node *Node) bool, nodes []*Node) []*Node {
	ret := make([]*Node, 0, len(nodes))
	for _, node := range nodes {
		if f(node) {
			ret = append(ret, node)
		}
	}
	return ret
}

func negate(f func(node *Node) bool) func(node *Node) bool {
	return func(node *Node) bool {
		return !f(node)
	}
}

func reverse(nodes []*Node) []*Node {
	l := len(nodes)
	ret := make([]*Node, l)
	for i, node := range nodes {
		ret[l-i-1] = node
	}
	return ret
}

// func IsOnStack(u, v, stk) {}
// func Stack2string(stk) {}

func Ext(u Combinator, v []*Node, stk []*Pair) []*Pair {
	return append(stk, &Pair{
		combinator: u,
		toks:       v,
	})
}

type Parser func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node)
type Combinator func() Parser

func ApplyCheck(combinator Combinator, toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
	return combinator()(toks, Ext(combinator, toks, stk), ctx)
}

// @seq
func AtSeq(cs ...Combinator) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			var loop func(cs []Combinator, toks []*Node, nodes []*Node) ([]*Node, []*Node)
			loop = func(cs []Combinator, toks []*Node, nodes []*Node) ([]*Node, []*Node) {
				if len(cs) == 0 {
					return nodes, toks
				}
				if t, r := ApplyCheck(cs[0], toks, stk, ctx); t == nil {
					return nil, nil
				} else {
					return loop(cs[1:], r, append(nodes, t...))
				}
			}
			return loop(cs, toks, make([]*Node, 0))
		}
	}
}

// @...
// 移除 phantoms
func AtDot(cs ...Combinator) Combinator {
	parser := AtSeq(cs...)()
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if t, r := parser(toks, stk, ctx); t == nil {
				return nil, nil
			} else {
				return filter(negate(IsPhantom), t), r
			}
		}
	}
}

// @or
func AtOr(cs ...Combinator) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			var loop func(cs []Combinator) ([]*Node, []*Node)
			loop = func(cs []Combinator) ([]*Node, []*Node) {
				if len(cs) == 0 {
					return nil, nil
				}
				if t, r := ApplyCheck(cs[0], toks, stk, ctx); t == nil {
					return loop(cs[1:])
				} else {
					return t, r
				}
			}
			return loop(cs)
		}
	}
}

// @=
func AtEq(tp string, cs ...Combinator) Combinator {
	parser := AtSeq(cs...)()
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if t, r := parser(toks, stk, ctx); t == nil {
				return nil, nil
			} else if tp == "" {
				return filter(negate(IsPhantom), t), r
			} else if l := len(t); l == 0 {
				return []*Node{
					{
						Type:  tp,
						Start: toks[0].Start,
						End:   toks[0].Start,
					},
				}, r
			} else {
				return []*Node{
					{
						Type:  tp,
						Start: t[0].Start,
						End:   t[l-1].End,
						Elts:  filter(negate(IsPhantom), t),
					},
				}, r
			}
		}
	}
}

// @*
func AtStar(cs ...Combinator) Combinator {
	parser := AtDot(cs...)()
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			var loop func(toks []*Node, nodes []*Node) ([]*Node, []*Node)
			loop = func(toks []*Node, nodes []*Node) ([]*Node, []*Node) {
				if len(toks) == 0 {
					return nodes, make([]*Node, 0)
				} else {
					if t, r := parser(toks, stk, ctx); t == nil {
						return nodes, toks
					} else {
						return loop(r, append(nodes, t...))
					}
				}
			}
			return loop(toks, make([]*Node, 0))
		}
	}
}

// @*^
func AtStar_(c Combinator) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			var loop func(toks []*Node, nodes []*Node) ([]*Node, []*Node)
			loop = func(toks []*Node, nodes []*Node) ([]*Node, []*Node) {
				if len(toks) == 0 {
					return nodes, make([]*Node, 0)
				}
				if t, r := c()(toks, stk, ctx); t == nil {
					return nodes, toks
				} else {
					return loop(r, append(nodes, t...))
				}
			}
			return loop(toks, make([]*Node, 0))
		}
	}
}

// @+
func AtAdd(c Combinator) Combinator {
	return AtDot(c, AtStar(c))
}

// @?
func AtWhy(cs ...Combinator) Combinator {
	return AtOr(AtDot(cs...), _none)
}

// @!
func AtFail(cs ...Combinator) Combinator {
	parser := AtDot(cs...)()
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if t, _ := parser(toks, stk, ctx); t == nil {
				return []*Node{toks[0]}, toks[1:]
			} else {
				return nil, nil
			}
		}
	}
}

// @!^
func AtFail_(c Combinator) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if t, _ := c()(toks, stk, ctx); t == nil {
				return []*Node{toks[0]}, toks[1:]
			} else {
				return nil, nil
			}
		}
	}
}

// @and
func AtAnd(cs ...Combinator) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			var loop func(cs []Combinator, res [][]*Node) ([]*Node, []*Node)
			loop = func(cs []Combinator, res [][]*Node) ([]*Node, []*Node) {
				if len(cs) == 0 {
					return res[0], res[1]
				}
				if t, r := ApplyCheck(cs[0], toks, stk, ctx); t == nil {
					return nil, nil
				} else {
					return loop(cs[1:], append([][]*Node{t, r}, res...))
				}
			}
			return loop(cs, make([][]*Node, 0))
		}
	}
}

// 跳过匹配的
// $glob
func _glob(cs ...Combinator) Combinator {
	parser := AtDot(cs...)()
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if t, r := parser(toks, stk, ctx); t == nil {
				return nil, nil
			} else {
				return make([]*Node, 0), r
			}
		}
	}
}

// 跳过单个匹配的
// $glob^
func _glob_(c Combinator) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if t, r := c()(toks, stk, ctx); t == nil {
				return nil, nil
			} else {
				return make([]*Node, 0), r
			}
		}
	}
}

// 将匹配当的作为分隔符
// $phantom
func _phantom(cs ...Combinator) Combinator {
	parser := AtDot(cs...)()
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if t, r := parser(toks, stk, ctx); t == nil {
				return nil, nil
			} else if l := len(t); l == 0 {
				return make([]*Node, 0), r
			} else {
				return []*Node{
					{
						Type:  PhantomType,
						Start: t[0].Start,
						End:   t[l-1].End,
					},
				}, r
			}
		}
	}
}

// $fail
func _fail() Parser {
	return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
		return nil, nil
	}
}

// $none
func _none() Parser {
	return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
		return make([]*Node, 0), toks
	}
}

// $pred
func _pred(proc func(*Node) bool) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if len(toks) == 0 {
				return nil, nil
			} else if proc(toks[0]) {
				return []*Node{toks[0]}, toks[1:]
			} else {
				return nil, nil
			}
		}
	}
}

// $eof
var _eof = _glob(_pred(func(t *Node) bool {
	return t.Type == EofType
}))

// $$
func __(s string) Combinator {
	return _pred(func(x *Node) bool {
		return IsTokenType(x) && x.Text == s
	})
}

// @_
func At_(s string) Combinator {
	return _glob(__(s))
}

// @~
func AtSkip(s string) Combinator {
	return _phantom(__(s))
}

func Join(cs []Combinator, sep Combinator) []Combinator {
	if len(cs) == 1 {
		return cs
	} else {
		return append([]Combinator{cs[0], sep}, Join(cs[1:], sep)...)
	}
}

// @.@
func AtDotAt(c, sep Combinator) Combinator {
	return AtDot(c, AtStar(AtDot(sep, c)))
}

// 后缀表达式
// @postfix
func AtPostfix(tp string, c, op Combinator) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if t, r := ApplyCheck(AtDot(c, AtAdd(op)), toks, stk, ctx); t == nil {
				return nil, nil
			} else {
				return []*Node{MakePostfix(tp, t)}, r
			}
		}
	}
}

func MakePostfix(tp string, ls []*Node) *Node {
	var loop func([]*Node, *Node) *Node
	loop = func(ls []*Node, ret *Node) *Node {
		if len(ls) == 0 {
			return ret
		} else {
			e := &Node{
				Type:  tp,
				Start: ret.Start,
				End:   ls[0].End,
				Elts:  []*Node{ret, ls[0]},
			}
			return loop(ls[1:], e)
		}
	}
	return loop(ls[1:], ls[0])
}

// 前缀表达式
// @prefix
func AtPrefix(tp string, c, op Combinator) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if t, r := ApplyCheck(AtDot(AtAdd(op), c), toks, stk, ctx); t == nil {
				return nil, nil
			} else {
				return []*Node{MakePrefix(tp, t)}, r
			}
		}
	}
}

func MakePrefix(tp string, ls []*Node) *Node {
	if len(ls) == 1 {
		return ls[0]
	} else {
		tail := MakePrefix(tp, ls[1:])
		return &Node{
			Type:  tp,
			Start: ls[0].Start,
			End:   tail.End,
			Elts:  []*Node{ls[0], tail},
		}
	}
}

// @infix
func AtInfix(tp string, c, op Combinator, associativity string) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			var loop func([]*Node, []*Node) ([]*Node, []*Node)
			loop = func(rest []*Node, ret []*Node) ([]*Node, []*Node) {
				if tc, rc := ApplyCheck(AtSeq(c), rest, stk, ctx); tc == nil {
					if lc := len(ret); lc < 3 {
						return nil, nil
					} else {
						fields := ret[:lc-1]
						constr := constrExpL
						if associativity == "right" {
							constr = constrExpR
						}
						return []*Node{constr(tp, fields)}, append([]*Node{ret[lc-1]}, rest...)
					}
				} else {
					if top, rop := ApplyCheck(AtSeq(op), rc, stk, ctx); top == nil {
						if lc := len(ret); lc < 2 {
							return nil, nil
						} else {
							fields := append(ret, tc...)
							constr := constrExpL
							if associativity == "right" {
								constr = constrExpR
							}
							return []*Node{constr(tp, fields)}, rc
						}
					} else {
						return loop(rop, append(ret, append(tc, top...)...))
					}
				}
			}
			return loop(toks, make([]*Node, 0))
		}
	}
}

// @infix-left
func AtInfixLeft(tp string, c, op Combinator) Combinator {
	return AtInfix(tp, c, op, "left")
}

// @infix-right
func AtInfixRight(tp string, c, op Combinator) Combinator {
	return AtInfix(tp, c, op, "right")
}

func constrExpL(tp string, fields []*Node) *Node {
	var loop func([]*Node, *Node) *Node
	loop = func(fields []*Node, ret *Node) *Node {
		if len(fields) == 0 {
			return ret
		} else {
			e := &Node{
				Type:  tp,
				Start: ret.Start,
				End:   fields[1].End,
				Elts:  []*Node{ret, fields[0], fields[1]},
			}
			return loop(fields[2:], e)
		}
	}
	return loop(fields[1:], fields[0])
}

func constrExpR(tp string, fields []*Node) *Node {
	fields = reverse(fields)
	var loop func([]*Node, *Node) *Node
	loop = func(fields []*Node, ret *Node) *Node {
		if len(fields) == 0 {
			return ret
		} else {
			e := &Node{
				Type:  tp,
				Start: fields[1].Start,
				End:   ret.End,
				Elts:  []*Node{fields[1], fields[0], ret},
			}
			return loop(fields[2:], e)
		}
	}
	return loop(fields[1:], fields[0])
}

// :: 加一层缓存
func CC(c Combinator) Combinator {
	return func() Parser {
		return func(toks []*Node, stk []*Pair, ctx interface{}) ([]*Node, []*Node) {
			if cache, ok := ctx.(map[string][][]*Node); !ok {
				return c()(toks, stk, ctx)
			} else if t, r := getCache(cache, c, toks); t != nil {
				return t, r
			} else {
				nt, nr := c()(toks, stk, ctx)
				setCache(cache, c, toks, nt, nr)
				return nt, nr
			}
		}
	}
}

func getCache(cache map[string][][]*Node, c Combinator, toks []*Node) ([]*Node, []*Node) {
	key := getCacheKey(c, toks)
	if res, ok := cache[key]; !ok {
		return nil, nil
	} else {
		return res[0], res[1]
	}
}

func setCache(cache map[string][][]*Node, c Combinator, toks, nt, nr []*Node) {
	key := getCacheKey(c, toks)
	cache[key] = [][]*Node{nt, nr}
}

func getCacheKey(c Combinator, toks []*Node) string {
	name := functionName(c)
	runes := make([]rune, 0, len(toks)*100)
	for _, tok := range toks {
		runes = append(runes, tok.Text...)
	}
	res, _ := json.Marshal([]string{name, string(runes)})
	return res
}

func functionName(i interface{}, seps ...rune) string {
	// 获取函数名称
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	// 用 seps 进行分割
	fields := strings.FieldsFunc(fn, func(sep rune) bool {
		for _, s := range seps {
			if sep == s {
				return true
			}
		}
		return false
	})

	if size := len(fields); size > 0 {
		return fields[size-1]
	}
	return ""
}

func Eval(c Combinator, toks []*Node) ([]*Node, []*Node) {
	cache := make(map[string][][]*Node, len(toks))
	return c()(toks, make([]*Pair, 0), cache)
}

var (
	B = map[string]func(...Combinator) Combinator{
		"@or":      AtOr,
		"@and":     AtAnd,
		"@!":       AtFail,
		"@seq":     AtSeq,
		"@*":       AtStar,
		"@...":     AtDot,
		"@?":       AtWhy,
		"$glob":    _glob,
		"$phantom": _phantom,
	}
	O = map[string]func(Combinator) Combinator{
		"@+":     AtAdd,
		"@*^":    AtStar_,
		"@!^":    AtFail_,
		"$glob^": _glob_,
		"::":     CC,
	}
	T = map[string]func(string, ...Combinator) Combinator{
		"@=": AtEq,
	}
	S = map[string]func(string) Combinator{
		"@~": AtSkip,
		"$$": __,
		"@_": At_,
	}
	C = map[string]Combinator{
		"$fail": _fail,
		"$none": _none,
		"$eof":  _eof,
	}
	P = map[string]func(func(*Node) bool) Combinator{
		"$pred": _pred,
	}
	J = map[string]func(c, sep Combinator) Combinator{
		"@.@": AtDotAt,
	}
	F = map[string]func(tp string, c, op Combinator) Combinator{
		"@prefix":      AtPrefix,
		"@postfix":     AtPostfix,
		"@infix-left":  AtInfixLeft,
		"@infix-right": AtInfixRight,
	}
)
