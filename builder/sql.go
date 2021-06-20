package builder

import (
	"github.com/Aiyane/parsec-go/parser"
	"strings"
)

func Node(node *parser.Node) string {
	switch node.Type {
	case "select":
		return Select(node)
	case "having":
		return Having(node)
	case "group":
		return Group(node)
	case "where":
		return Where(node)
	case "on":
		return On(node)
	case "join":
		return Join(node)
	case "as":
		return As(node)
	case "from":
		return From(node)
	case "field":
		return Field(node)
	case "func":
		return Func(node)
	case "desc":
		return Desc(node)
	case "aes":
		return Aes(node)
	case "order":
		return Order(node)
	case "limit":
		return Limit(node)
	default:
		return node.Text
	}
}

func Nodes(nodes []*parser.Node) []string {
	ss := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ss = append(ss, Node(node))
	}
	return ss
}

func Select(node *parser.Node) string {
	ss := Nodes(node.Elts[1:])
	s := Node(node.Elts[0])
	return "(SELECT " + strings.Join(ss, ", ") + s + ")"
}

func Having(node *parser.Node) string {
	ss := Nodes(node.Elts[1:])
	s := Node(node.Elts[0])
	return s + " HAVING " + strings.Join(ss, " AND ")
}

func Group(node *parser.Node) string {
	ss := Nodes(node.Elts[1:])
	s := Node(node.Elts[0])
	return s + " GROUP BY " + strings.Join(ss, ", ")
}

func Where(node *parser.Node) string {
	ss := Nodes(node.Elts[1:])
	s := Node(node.Elts[0])
	return s + " WHERE " + strings.Join(ss, " AND ")
}

func On(node *parser.Node) string {
	ss := Nodes(node.Elts[1:])
	s := Node(node.Elts[0])
	return s + " ON " + strings.Join(ss, " AND ")
}

func Join(node *parser.Node) string {
	ss := Nodes(node.Elts[1:])
	s := Node(node.Elts[0])
	return s + " JOIN " + strings.Join(ss, ", ")
}

func As(node *parser.Node) string {
	s1 := Node(node.Elts[0])
	s2 := Node(node.Elts[1])
	return s1 + " AS " + s2
}

func From(node *parser.Node) string {
	s := Node(node.Elts[0])
	return " FROM " + s
}

func Field(node *parser.Node) string {
	s1 := Node(node.Elts[0])
	s2 := Node(node.Elts[1])
	return s1 + "." + s2
}

func Func(node *parser.Node) string {
	f := Node(node.Elts[0])
	ss := Nodes(node.Elts[1:])
	if f == "=" || f == "==" || f == ">" || f == "<" || f == "!=" || f == "<=" ||
		f == ">=" || f == "+" || f == "-" || f == "*" || f == "/" || f == "in" ||
		f == "and" || f == "or" {
		return strings.Join(ss, " "+f+" ")
	}
	return f + "(" + strings.Join(ss, ",") + ")"
}

func Desc(node *parser.Node) string {
	s := Node(node.Elts[0])
	return s + " DESC"
}

func Aes(node *parser.Node) string {
	s := Node(node.Elts[0])
	return s + " AES"
}

func Order(node *parser.Node) string {
	ss := Nodes(node.Elts[1:])
	s := Node(node.Elts[0])
	return s + " ORDER BY " + strings.Join(ss, ", ")
}

func Limit(node *parser.Node) string {
	s1 := Node(node.Elts[0])
	s2 := Node(node.Elts[1])
	s3 := Node(node.Elts[2])
	return s1 + " OFFSET " + s2 + " LIMIT " + s3
}
