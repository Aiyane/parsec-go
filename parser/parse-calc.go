package parser

func SetCalcParameters() {
	SetWhiteSpace(" ", "\n")
	SetDelims("(", ")", "[", "]")
	SetOperators("==", "!=", ">=", "<=", "&&", "||", ">>", "<<", "++", "--",
				"+", "-", "*", "/", "%", "~", "!", ":", "?", ">", "<", "|", "^", "&")
	SetQuotationMarks("\"", "'")
}

// utility for constructing operators
func op(s string) Combinator { return S["$$"](s) }

// ?:	 Ternary conditional
// --------------------------------------------
//
//	 conditionalExpression ::
//		logicalOrExpression `?` conditionalExpression `:` conditionalExpression
//		| logicalOrExpression
func conditionalExpression() Parser {
	return B["@or"](
		T["@="](Conditional,
			O["::"](logicalOrExpression),
			S["@~"]("?"), O["::"](conditionalExpression),
			S["@~"](":"), O["::"](conditionalExpression)),

		O["::"](logicalOrExpression))()
}

// ||	 Logical OR
// --------------------------------------------
//
//	 logicalOrExpression ::
//		logicalAndExpression `||` logicalAndExpression
//		| logicalAndExpression
func logicalOrExpression() Parser {
	return B["@or"](
		F["@infix-left"](LogicalOR,
			O["::"](logicalAndExpression),
			op("||")),

		O["::"](logicalAndExpression))()
}

// &&	 Logical AND
// --------------------------------------------
//
//	 logicalAndExpression ::
//		bitwiseOrExpression `&&` bitwiseOrExpression
//		| bitwiseOrExpression
func logicalAndExpression() Parser {
	return B["@or"](
		F["@infix-left"](LogicalAND,
			O["::"](bitwiseOrExpression),
			op("&&")),

		O["::"](bitwiseOrExpression))()
}

// |	 Bitwise OR (inclusive or)
// --------------------------------------------
//
//	 bitwiseOrExpression ::
//		bitwiseXorExpression `|` bitwiseXorExpression
//		| bitwiseXorExpression
func bitwiseOrExpression() Parser {
	return B["@or"](
		F["@infix-left"](BitwiseOR,
			O["::"](bitwiseXorExpression),
			op("|")),

		O["::"](bitwiseXorExpression))()
}

// ^	 Bitwise XOR (exclusive or)
// --------------------------------------------
//
//	 bitwiseXorExpression ::
//		bitwiseAndExpression `^` bitwiseAndExpression
//		| bitwiseAndExpression
func bitwiseXorExpression() Parser {
	return B["@or"](
		F["@infix-left"](BitwiseXOR,
			O["::"](bitwiseAndExpression),
			op("^")),

		O["::"](bitwiseAndExpression))()
}

// &	 Bitwise AND
// --------------------------------------------
//
//	 bitwiseAndExpression ::
//		equalityExpression `&` equalityExpression
//		| equalityExpression
func bitwiseAndExpression() Parser {
	return B["@or"](
		F["@infix-left"](BitwiseAND,
			O["::"](equalityExpression),
			op("&")),

		O["::"](equalityExpression))()
}

// equality
// --------------------------------------------
//
//	 equalityExpression ::
//		relationalExpression `==` relationalExpression
//		| relationalExpression `!=` relationalExpression
//		| relationalExpression
func equalityExpression() Parser {
	return B["@or"](
		F["@infix-left"](Equality,
			O["::"](relationalExpression),
			O["::"](equalityOperator)),

		O["::"](relationalExpression))()
}

var equalityOperator = B["@or"](op("=="), op("!="))

// relational
// --------------------------------------------
//
//	 relationalExpression ::
//		BitwiseShiftExpression `<` BitwiseShiftExpression
//		| BitwiseShiftExpression `<=` BitwiseShiftExpression
//		| BitwiseShiftExpression `>` BitwiseShiftExpression
//		| BitwiseShiftExpression `>=` BitwiseShiftExpression
//		| BitwiseShiftExpression
func relationalExpression() Parser {
	return B["@or"](
		F["@infix-left"](Relational,
			O["::"](BitwiseShiftExpression),
			O["::"](relationalOperator)),

		O["::"](BitwiseShiftExpression))()
}

var relationalOperator = B["@or"](op("<"), op("<="), op(">"), op(">="))

// bitwise shift
// --------------------------------------------
//
//	 BitwiseShiftExpression ::
//		additiveExpression `<<` additiveExpression
//		| additiveExpression `>>` additiveExpression
//		| additiveExpression
func BitwiseShiftExpression() Parser {
	return B["@or"](
		F["@infix-left"](BitwiseShift,
			O["::"](additiveExpression),
			O["::"](BitwiseShiftOperator)),

		O["::"](additiveExpression))()
}

var BitwiseShiftOperator = B["@or"](op("<<"), op(">>"))

// additive
// --------------------------------------------
//
//	 additiveExpression ::
//		multiplicativeExpression `+` multiplicativeExpression
//		| multiplicativeExpression `-` multiplicativeExpression
//		| multiplicativeExpression
func additiveExpression() Parser {
	return B["@or"](
		F["@infix-left"](Additive,
			O["::"](multiplicativeExpression),
			O["::"](additiveOperator)),

		O["::"](multiplicativeExpression))()
}

var additiveOperator = B["@or"](op("+"), op("-"))

// multiplicative
// --------------------------------------------
//
//	 multiplicativeExpression ::
//		prefixExpression `*` prefixExpression
//		| prefixExpression `/` prefixExpression
//		| prefixExpression `%` prefixExpression
//		| prefixExpression
func multiplicativeExpression() Parser {
	return B["@or"](
		F["@infix-left"](Multiplicative,
			O["::"](prefixExpression),
			O["::"](multiplicativeOperator)),

		O["::"](prefixExpression))()
}

var multiplicativeOperator = B["@or"](op("*"), op("/"), op("%"))

// prefix
// --------------------------------------------
//
//	 prefixExpression ::
//		`++` postfixExpression
//		| `--` postfixExpression
//		| `+` postfixExpression
//		| `-` postfixExpression
//		| `~` postfixExpression
//		| `!` postfixExpression
//		| postfixExpression
func prefixExpression() Parser {
	return B["@or"](
		F["@prefix"](Prefix,
			O["::"](postfixExpression),
			O["::"](prefixOperator)),

		O["::"](postfixExpression))()
}

var prefixOperator = B["@or"](op("++"), op("--"), op("+"), op("-"), op("~"), op("!"))

// postfix
// --------------------------------------------
//
//	 postfixExpression ::
//		primaryExpression `++`
//		| primaryExpression `--`
//		| primaryExpression
func postfixExpression() Parser {
	return B["@or"](
		F["@postfix"](Postfix,
			O["::"](primaryExpression),
			O["::"](postfixOperator)),

		O["::"](primaryExpression))()
}

var postfixOperator = B["@or"](op("++"), op("--"))

// primary
// --------------------------------------------
//
//	 primaryExpression ::
//		literal
//		| `(` conditionalExpression `)`
func primaryExpression() Parser {
	return B["@or"](literal, T["@="](Expression, S["@~"]("("), O["::"](conditionalExpression), S["@~"](")")))()
}

//	 literal ::
//		boolLiteral
//		| stringLiteral
//		| intLiteral
//		| floatLiteral
var literal = B["@or"](boolLiteral, stringLiteral, intLiteral, floatLiteral)

//	 boolLiteral ::
//		`true` | `false`
var boolLiteral = T["@="](Bool, B["@or"](S["$$"]("true"), S["$$"]("false")))

var stringLiteral = P["$pred"](IsStrType)

var intLiteral = T["@="](Int, P["$pred"](func(node *Node) bool {
	return IsTokenType(node) && IsNumeral(node.Text) && !HasDot(node.Text)
}))

var floatLiteral = T["@="](Float, P["$pred"](func(node *Node) bool {
	return IsTokenType(node) && IsNumeral(node.Text) && HasDot(node.Text)
}))

var identifier = P["$pred"](func(node *Node) bool { return IsTokenType(node) && IsId(node.Text) })

const (
	Conditional    string = "conditional-expression"
	LogicalOR      string = "logical-or"
	LogicalAND     string = "logical-and"
	BitwiseOR      string = "bitwise-or"
	BitwiseXOR     string = "bitwise-xor"
	BitwiseAND     string = "bitwise-and"
	Equality       string = "equality"
	Relational     string = "relational"
	BitwiseShift   string = "bitwise-shift"
	Additive       string = "additive"
	Multiplicative string = "multiplicative"
	Prefix         string = "prefix"
	Postfix        string = "postfix"
	Expression     string = "expression"
	Bool           string = "bool"
	Int            string = "int"
	Float          string = "float"
)

func IsNumeral(s []rune) bool {
	if len(s) == 0 {
		return false
	}
	if IsDigit(s[0:1]) {
		return true
	}
	return false
}

func HasDot(s []rune) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}
	return false
}

func ParseCalc(s string) []*Node {
	SetCalcParameters()
	t, _ := Eval(B["@*"](O["::"](conditionalExpression)), Scan(s))
	return t
}
