package parser

import (
	"os"
	"fmt"
	"asm/token"
)

type parser struct {
	toks []token.Token
	tok *token.Token
	cur int
}

type Directive struct {
	Kind token.Kind
	Arg *token.Token
}

type Label struct {
	Name *token.Token
}

type Instruction struct {
	Kind token.Kind
	Args []*token.Token
}

type Stmt interface{}

func Parse(toks []token.Token) []Stmt {
	ss := make([]Stmt, 0, 512)
	p := parser{toks, &toks[0], 0}

	for p.tok.Kind != token.EOF {
		var s Stmt

		switch p.tok.Kind {
		case token.Dot:
			s = p.parseDirective()
		case token.Sym:
			s = p.parseLabel()
		case token.LF:
			p.advance()
			continue
		default:
			s = p.parseInstruction()
		}

		ss = append(ss, s)
	}

	return ss
}

func (p *parser) advance() *token.Token {
	if p.tok.Kind == token.EOF {
		return p.tok
	}

	prev := p.tok
	p.cur++
	p.tok = &p.toks[p.cur]

	return prev
}

func (p *parser) consume(kinds ...token.Kind) *token.Token {
	var msg string

	for i, k := range kinds {
		if p.tok.Kind == k {
			return p.advance()
		}
		if i > 0 {
			msg += ", "
		}
		msg += k.String()
	}

	fmt.Fprintf(os.Stderr, "%s: expected %s but got %s\n", p.tok.Pos, msg, p.tok.Kind)
	os.Exit(1)
	panic("i just want to kill the process")
}

func (p *parser) consumeReg() *token.Token {
	if !p.tok.Kind.IsRegister() {
		fmt.Fprintf(os.Stderr, "%s: expected register but got %s\n", p.tok.Pos, p.tok.Kind)
		os.Exit(1)
	}
	return p.advance()
}

func (p *parser) parseDirective() Stmt {
	p.consume(token.Dot)

	dir := p.advance()
	var arg *token.Token

	switch dir.Kind {
	case token.Extern, token.Global:
		arg = p.consume(token.Sym)
	case token.Byte:
		arg = p.consume(token.Num, token.Char)
	case token.Word, token.Skip:
		arg = p.consume(token.Num)
	case token.Ascii:
		arg = p.consume(token.Str)
	default:
		fmt.Fprintf(os.Stderr, "%s: expected directive but got %s\n", dir.Pos, dir.Kind)
		os.Exit(1)
	}

	p.consume(token.LF)

	return Directive{dir.Kind, arg}
}

func (p *parser) parseLabel() Stmt {
	sym := p.consume(token.Sym)
	p.consume(token.Colon)
	p.consume(token.LF)

	return Label{sym}
}

func (p *parser) parseInstruction() Stmt {
	op := p.advance()
	args := make([]*token.Token, 0, 8)

	switch op.Kind {
	case token.Mov:
		arg1 := p.advance()
		if !arg1.Kind.IsRegister() && arg1.Kind != token.Sym && arg1.Kind != token.Num && arg1.Kind != token.Char {
			fmt.Fprintf(os.Stderr, "%s: expected register, symbol, number or character but got %s\n", op.Pos, arg1.Kind)
			os.Exit(1)
		}
		p.consume(token.Comma)
		arg2 := p.consumeReg()
		args = append(args, arg1, arg2)
	case token.Movb, token.Movze, token.Movse, token.Wr, token.Wrb, token.Rd, token.Rdb, token.Add, token.Addb,
			token.Sub, token.Subb, token.Cmp, token.Cmpb:
		arg1 := p.consumeReg()
		p.consume(token.Comma)
		arg2 := p.consumeReg()
		args = append(args, arg1, arg2)
	case token.Jmp, token.Jz, token.Je, token.Jnz, token.Jne, token.Jc, token.Jb, token.Jnc, token.Jae, token.Js,
			token.Jns, token.Jo, token.Jno, token.Jbe, token.Ja, token.Jl, token.Jge, token.Jle, token.Jg, token.Call:
		arg1 := p.consume(token.Sym)
		args = append(args, arg1)
	case token.Push, token.Pop:
		arg1 := p.consumeReg()
		args = append(args, arg1)
	case token.Halt, token.Ret, token.Syscall: // art: 0 args
	default:
		fmt.Fprintf(os.Stderr, "%s: expected instruction but got %s\n", op.Pos, op.Kind)
		os.Exit(1)
	}

	p.consume(token.LF)

	return Instruction{op.Kind, args}
}
