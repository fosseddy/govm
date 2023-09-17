/*
program = (extern_decl|global_decl|data_label|text_label)* EOF

extern_decl = "." "extern" symbol ("," symbol)*
global_decl = "." "global" symbol ("," symbol)*

data_label = label (directive LF)*
text_label = label (mnemonic LF|label)*

label = symbol ":" LF

directive = "." ("byte"   (number|char)
				|"word"    number
				|"string"  string
				|"skip"    number)

mnemonic = "halt"
		 | "mov" "b"? (reg|number|char) "," reg
	     | "movze" reg "," reg
	     | "movse" reg "," reg
	     | "st"  "b"? reg "," reg
	     | "ld"  "b"? reg "," reg
	     | "add" "b"? reg "," reg
	     | "sub" "b"? reg "," reg
	     | "cmp" "b"? reg "," reg
	     | "j" ("mp"|"z" |"e"
		 	   |"nz"|"ne"|"c"
			   |"b" |"nc"|"ae"
			   |"s" |"ns"|"o"
			   |"no"|"be"|"a"
			   |"l" |"ge"|"le"
			   |"g") symbol
	     | "push" reg
	     | "pop"  reg
	     | "call" symbol
	     | "ret"
	     | "syscall"

reg    = "r" ("0".."13"|"sp"|"bp")
number = digit+
symbol = letter (letter|digit)*
string = "<any char except ">+"
char   = '<any char except '>'

letter = "a".."z"|"A".."Z"|"_"
digit  = "0".."9"

*/

package main

import (
	"os"
	"fmt"
	"asm/scanner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Provide file to assemble")
		os.Exit(1)
	}

	s := scanner.New(os.Args[1])
	toks := s.Parse()

	fmt.Println(toks)
}
