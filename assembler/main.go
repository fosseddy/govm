/*
program = (label|mnemonic|directive)* EOF

label = symbol ":" LF

directive = "." ("global" symbol
				|"extern" symbol
				|"byte"  (number|char)
				|"word"   number
				|"ascii"  string
				|"skip"   number)

mnemonic = "halt"
		 | "mov" "b"? reg "," reg
		 | "movi" (number|char|symbol) "," reg
	     | "movze" reg "," reg
	     | "movse" reg "," reg
	     | "wr"  "b"? reg "," reg
	     | "rd"  "b"? reg "," reg
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

/*
Object file

Header
  nsyms - 2 bytes
  nrels - 2 bytes

Code
  len  - 2 bytes
  code - len bytes

Symbols
  kind   - 1 byte
  idx    - 2 bytes
  addr   - 2 bytes
  nlabel - 2 bytes
  label  - nlabel bytes

Relocations
  loc    - 2 bytes
  symidx - 2 bytes
 */

package main

import (
	"os"
	"fmt"
	"bytes"
	"encoding/binary"
	"asm/parser"
	"asm/scanner"
	"asm/token"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Provide file to assemble")
		os.Exit(1)
	}

	toks := scanner.Scan(os.Args[1])
	stmts := parser.Parse(toks)

	st := symtab{}
	st.populate(stmts)

	code := new(bytes.Buffer)

	for _, s := range stmts {
		switch s := s.(type) {
		case parser.Directive:
			switch s.Kind {
			case token.Byte:
				v := uint8(s.Arg.Value)
				binary.Write(code, binary.LittleEndian, v)
			case token.Word:
				v := uint16(s.Arg.Value)
				binary.Write(code, binary.LittleEndian, v)
			case token.Ascii:
				v := []byte(s.Arg.Lex[1:len(s.Arg.Lex)-1])
				binary.Write(code, binary.LittleEndian, v)
			case token.Skip:
				v := make([]byte, s.Arg.Value)
				binary.Write(code, binary.LittleEndian, v)
			}
		case parser.Instruction:
			v := encodeInstruction(&s, st)
			binary.Write(code, binary.LittleEndian, v)
		}
	}

	f, _ := os.Create("out.vm")
	binary.Write(f, binary.LittleEndian, uint16(st["_start"].addr))
	f.Write(code.Bytes())
}

func encodeOp(kind token.Kind) uint8 {
	switch kind {
	case token.Halt:
		return 0
	case token.Mov:
		return 1
	case token.Movb:
		return 2
	case token.Movi:
		return 3
	case token.Movze:
		return 4
	case token.Movse:
		return 5
	case token.Wr:
		return 6
	case token.Wrb:
		return 7
	case token.Rd:
		return 8
	case token.Rdb:
		return 9
	case token.Add:
		return 10
	case token.Addb:
		return 11
	case token.Sub:
		return 12
	case token.Subb:
		return 13
	case token.Cmp:
		return 14
	case token.Cmpb:
		return 15
	case token.Jmp, token.Jz, token.Je, token.Jnz, token.Jne, token.Jc, token.Jb, token.Jnc, token.Jae, token.Js,
			token.Jns, token.Jo, token.Jno, token.Jbe, token.Ja, token.Jl, token.Jge, token.Jle, token.Jg:
		return 16
	case token.Push:
		return 17
	case token.Pop:
		return 18
	case token.Call:
		return 19
	case token.Ret:
		return 20
	case token.Syscall:
		return 21
	}

	panic("unreachable")
}

func encodeReg(reg token.Kind) uint8 {
	switch reg {
	case token.R0:
		return 0
	case token.R1:
		return 1
	case token.R2:
		return 2
	case token.R3:
		return 3
	case token.R4:
		return 4
	case token.R5:
		return 5
	case token.R6:
		return 6
	case token.R7:
		return 7
	case token.R8:
		return 8
	case token.R9:
		return 9
	case token.R10:
		return 10
	case token.R11:
		return 11
	case token.R12:
		return 12
	case token.R13:
		return 13
	case token.Rsp:
		return 14
	case token.Rbp:
		return 15
	}

	panic("unreachable " + reg.String())
}

func encodeBranch(br token.Kind) uint8 {
	switch br {
	case token.Jmp:
		return 0
	case token.Jz, token.Je:
		return 1
	case token.Jnz, token.Jne:
		return 2
	case token.Jc, token.Jb:
		return 3
	case token.Jnc, token.Jae:
		return 4
	case token.Js:
		return 5
	case token.Jns:
		return 6
	case token.Jo:
		return 7
	case token.Jno:
		return 8
	case token.Jbe:
		return 9
	case token.Ja:
		return 10
	case token.Jl:
		return 11
	case token.Jge:
		return 12
	case token.Jle:
		return 13
	case token.Jg:
		return 14
	}

	panic("unreachable")
}

func encodeInstruction(inst *parser.Instruction, st symtab) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, encodeOp(inst.Kind))

	switch inst.Kind {
	case token.Mov, token.Movb, token.Movze, token.Movse, token.Wr, token.Wrb, token.Rd, token.Rdb, token.Add,
			token.Addb, token.Sub, token.Subb, token.Cmp, token.Cmpb:
		binary.Write(buf, binary.LittleEndian, encodeReg(inst.Args[0].Kind) << 4 | encodeReg(inst.Args[1].Kind))

	case token.Movi:
		binary.Write(buf, binary.LittleEndian, encodeReg(inst.Args[1].Kind))
		if inst.Args[0].Kind == token.Sym {
			binary.Write(buf, binary.LittleEndian, uint16(st[inst.Args[0].Lex].addr))
		} else {
			binary.Write(buf, binary.LittleEndian, uint16(inst.Args[0].Value))
		}

	case token.Jmp, token.Jz, token.Je, token.Jnz, token.Jne, token.Jc, token.Jb, token.Jnc, token.Jae, token.Js,
			token.Jns, token.Jo, token.Jno, token.Jbe, token.Ja, token.Jl, token.Jge, token.Jle, token.Jg:
		binary.Write(buf, binary.LittleEndian, encodeBranch(inst.Kind))
		binary.Write(buf, binary.LittleEndian, uint16(st[inst.Args[0].Lex].addr))

	case token.Push, token.Pop:
		binary.Write(buf, binary.LittleEndian, encodeReg(inst.Args[0].Kind))

	case token.Call:
		binary.Write(buf, binary.LittleEndian, uint16(st[inst.Args[0].Lex].addr))

	case token.Syscall, token.Ret, token.Halt: // art: 0 args
	default:
		panic("unreachable " + inst.Kind.String())
	}

	return buf.Bytes()
}

type symtab map[string]symbol

type symbol struct {
	kind symkind
	addr int
	pos token.Position
}

type symkind int
const (
	symlocal symkind = iota
	symglobal
	symextern
)

func (st symtab) populate(stmts []parser.Stmt) {
	addr := 0

	for _, s := range stmts {
		switch s := s.(type) {
		case parser.Directive:
			switch s.Kind {
			case token.Global:
				st[s.Arg.Lex] = symbol{symglobal, -1, s.Arg.Pos}
			case token.Extern:
				st[s.Arg.Lex] = symbol{symextern, -1, s.Arg.Pos}
			case token.Byte:
				addr++
			case token.Word:
				addr += 2
			case token.Ascii:
				addr += len(s.Arg.Lex) - 2 // art: -2 for string quotes
			case token.Skip:
				addr += s.Arg.Value
			default:
				panic("unreachable")
			}
		case parser.Label:
			newsym := symbol{symlocal, addr, s.Name.Pos}
			if sym, ok := st[s.Name.Lex]; ok {
				if sym.kind == symextern {
					fmt.Fprintf(os.Stderr, "%s: redefinition of external symbol\n", s.Name.Pos)
					os.Exit(1)
				}
				if sym.addr != -1 {
					fmt.Fprintf(os.Stderr, "%s: symbol %s already defined\n", s.Name.Pos, s.Name.Lex)
					os.Exit(1)
				}
				newsym.kind = sym.kind
			}
			st[s.Name.Lex] = newsym
		case parser.Instruction:
			switch s.Kind {
			case token.Halt, token.Ret, token.Syscall:
				addr += 1

			case token.Mov, token.Movb, token.Movze, token.Movse, token.Wr, token.Wrb, token.Rd, token.Rdb, token.Add,
					token.Addb, token.Sub, token.Subb, token.Cmp, token.Cmpb, token.Push, token.Pop:
				addr += 2

			case token.Call:
				addr += 3

			case token.Movi, token.Jmp, token.Jz, token.Je, token.Jnz, token.Jne, token.Jc, token.Jb, token.Jnc,
					token.Jae, token.Js, token.Jns, token.Jo, token.Jno, token.Jbe, token.Ja, token.Jl, token.Jge,
					token.Jle, token.Jg:
				addr += 4

			default:
				panic("unreachable")
			}
		default:
			panic("unreachable")
		}
	}

	for name, sym := range st {
		if sym.addr == -1 && sym.kind != symextern {
			fmt.Fprintf(os.Stderr, "%s: undefined symbol %s\n", sym.pos, name)
			os.Exit(1)
		}
	}
}
