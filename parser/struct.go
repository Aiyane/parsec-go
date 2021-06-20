package parser

const (
	CommentType   = "comment"
	PhantomType   = "phantom"
	TokenType     = "token"
	StrType       = "str"
	CharacterType = "character"
	NewlineType   = "newline"
	EofType       = "eof"
)

type Node struct {
	Type string
	Elts []*Node
	Text string
	Ctx  interface{}

	Start, End, Size int
}

func NewNode(t string, start, end int, elts []*Node, text string, size int, ctx interface{}) *Node {
	return &Node{
		Type:  t,
		Elts:  elts,
		Ctx:   ctx,
		Text:  text,
		Start: start,
		End:   end,
		Size:  size,
	}
}

func IsComment(n *Node) bool {
	return CommentType == n.Type
}

func IsPhantom(n *Node) bool {
	return PhantomType == n.Type
}

func IsTokenType(n *Node) bool {
	return TokenType == n.Type
}

func IsStrType(n *Node) bool {
	return StrType == n.Type
}

func IsCharacter(n *Node) bool {
	return CharacterType == n.Type
}

func IsNewline(n *Node) bool {
	return NewlineType == n.Type
}

var (
	left_recur_detection    = false
	delims                  = []string{"(", ")", "[", "]", "{", "}", "'", "`", ","}
	line_comment            = []string{";"}
	comment_start           = "#|"
	comment_end             = "|#"
	operators               = []string{}
	quotation_marks         = []string{"\"", "'"}
	lisp_char               = []string{"#\\", "?\\"}
	significant_whitespaces = []string{}
)

func SetDelims(x ...string) {
	delims = x
}

func SetLineComment(x ...string) {
	line_comment = x
}

func SetCommentStart(x string) {
	comment_start = x
}

func SetCommentEnd(x string) {
	comment_end = x
}

func SetOperators(x ...string) {
	operators = x
}

func SetQuotationMarks(x ...string) {
	quotation_marks = x
}

func SetLispChar(x ...string) {
	lisp_char = x
}

func SetSignificantWhitespaces(x ...string) {
	significant_whitespaces = x
}

func IsWhitespace(s string) bool {
	return s == "\t" || s == "\n" || s == "\v" || s == "\f" || s == "\r" || s == " "
}

func IsAlpha(s string) bool {
	return s == "a" || s == "b" || s == "c" || s == "d" || s == "e" || s == "f" || s == "g" ||
		s == "h" || s == "i" || s == "j" || s == "k" || s == "l" || s == "m" || s == "n" ||
		s == "o" || s == "p" || s == "q" || s == "r" || s == "s" || s == "t" || s == "u" ||
		s == "v" || s == "w" || s == "x" || s == "y" || s == "z" || s == "A" || s == "B" ||
		s == "C" || s == "D" || s == "E" || s == "F" || s == "G" || s == "H" || s == "I" ||
		s == "J" || s == "K" || s == "L" || s == "M" || s == "N" || s == "O" || s == "P" ||
		s == "Q" || s == "R" || s == "S" || s == "T" || s == "U" || s == "V" || s == "W" ||
		s == "X" || s == "Y" || s == "Z"
}

func IsDigit(s string) bool {
	return s == "0" || s == "1" || s == "2" || s == "3" || s == "4" || s == "5" || s == "6" ||
		s == "7" || s == "8" || s == "9"
}

func IsDelim(c string) bool {
	for _, s := range delims {
		if c == s {
			return true
		}
	}
	return false
}

func IsId(s string) bool {
	if len(s) == 0 {
		return false
	}
	if IsAlpha(s[0:1]) || "_" == s[0:1] {
		var loop func(i int) bool
		loop = func(i int) bool {
			if i >= len(s) {
				return true
			} else {
				c := s[i : i+1]
				if IsAlpha(c) || IsDigit(c) || "_" == c {
					return loop(i + 1)
				}
				return false
			}
		}
		return loop(1)
	}
	return false
}

func IsNumeral(s string) bool {
	if len(s) == 0 {
		return false
	}
	if IsDigit(s[0:1]) {
		return true
	}
	return false
}

func StartWith(s string, start int, prefix string) string {
	length := len(prefix)
	if length == 0 || len(s) < start+length {
		return ""
	}
	if s[start:start+length] == prefix {
		return prefix
	}
	return ""
}

func StartWithOneOf(s string, start int, prefixes []string) string {
	if len(prefixes) == 0 {
		return ""
	}
	if prefix := prefixes[0]; StartWith(s, start, prefix) != "" {
		return prefix
	}
	return StartWithOneOf(s, start, prefixes[1:])
}

func FindNext(s string, start int, pred func(string, int) bool) int {
	if len(s) <= start {
		return -1
	}
	if pred(s, start) {
		return start
	}
	return FindNext(s, 1+start, pred)
}

func FindDelim(s string, start int) string {
	return StartWithOneOf(s, start, delims)
}

func FindOperator(s string, start int) string {
	return StartWithOneOf(s, start, operators)
}
