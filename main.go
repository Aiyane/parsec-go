package main

import (
	"fmt"
	"github.com/Aiyane/parsec-go/builder"
	"github.com/Aiyane/parsec-go/parser"
)

func main() {
	s := `
(SELECT
	(LIMIT
		(ORDER
			(HAVING
				(GROUP 
					(WHERE
					   (ON (JOIN (FROM (AS tableA a))
								 (AS tableB b))
						   (= (. a filedA) (. b filedB)))
					   (= (. a filedA) (. b filedB))
			           (!= (. a filedB) (. b filedA)))
					(. a b))
				(> (. a b) 10))
			(DESC (. a user_id))
			(AES (. b id)))
		10 100)

	(AS (. a filedA) hello)
	(AS (. b filedB) world)
	(groupArray user_id)
	(distinct id)
	10
)
`
	nodes := parser.ParseSexp(s)
	if nodes == nil {
		fmt.Println("failed!")
	}
	nodes = parser.ParseSQL(s)
	if nodes == nil {
		fmt.Println("failed!")
	} else {
		sql := builder.Select(nodes[0])
		fmt.Println(sql)
	}

	s = `
(SELECT
	(WHERE
	   (FROM (AS tableA a))
	   (in (. a filedA) 
		   (SELECT (WHERE (FROM tableB)
						  (> id 10))
			       (. b user_id)))
	   (!= (. a filedB) (. b filedA)))

	(AS (. a filedA) hello)
	(AS (. b filedB) world)
	(groupArray user_id)
	(distinct id)
	10
)
`
	nodes = parser.ParseSQL(s)
	if nodes == nil {
		fmt.Println("failed!")
	} else {
		sql := builder.Select(nodes[0])
		fmt.Println(sql)
	}

	s = `2+3*num`
	nodes = parser.ParseCalc(s)
	if nodes == nil {
		fmt.Println("failed!")
	}
}
