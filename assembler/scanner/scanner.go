package scanner

import (
	"log"
	"fmt"
	"os"
	"strconv"
	"asm/token"
)

type Scanner struct {
	file string
	src []byte
	line int
	start int
	cur int
	ch byte
}

func New(file string) *Scanner {
	src, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	s := &Scanner{file: file, src: src, line: 1}
	if len(src) > 0 {
		s.ch = s.src[0]
	}

	return s
}

func (s *Scanner) Parse() []token.Token {
	toks := make([]token.Token, 0, 512)

	for {
		t := s.parseToken()
		toks = append(toks, t)
		if t.Kind == token.Eof {
			break
		}
	}

	return toks
}

func (s *Scanner) hasSrc() bool {
	return s.cur < len(s.src)
}

func (s *Scanner) advance() {
	s.cur++

	if !s.hasSrc() {
		return
	}

	s.ch = s.src[s.cur]
}

func (s *Scanner) next(ch byte) bool {
	next := s.cur + 1
	return next < len(s.src) && s.src[next] == ch
}

func (s *Scanner) makeToken(kind token.Kind) token.Token {
	tok := token.Token{
		Kind: kind,
		Lex: s.lexeme(),
		Pos: token.Position{s.file, s.line},
	}

	if tok.Kind == token.Num {
		val, err := strconv.Atoi(tok.Lex)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		tok.Value = val
	}

	return tok
}

func (s *Scanner) lexeme() string {
	return string(s.src[s.start:s.cur])
}

func (s *Scanner) parseToken() token.Token {
scanAgain:
	s.start = s.cur

	if !s.hasSrc() {
		return s.makeToken(token.Eof)
	}

	switch s.ch {
	case ' ', '\t', '\n', '\r':
		if s.ch == '\n' {
			s.line++
		}
		s.advance()
		goto scanAgain
	case '/':
		if s.next('/') {
			for s.ch != '\n' && s.hasSrc() {
				s.advance()
			}
			goto scanAgain
		}
		goto scanError
	
	case ':':
		s.advance()
		return s.makeToken(token.Colon)
	case ',':
		s.advance()
		return s.makeToken(token.Comma)
	case '.':
		s.advance()
		return s.makeToken(token.Dot)

	default:
		switch {
		case isLetter(s.ch):
			for isAlpha(s.ch) {
				s.advance()
			}
			kind := token.LookupKeyword(s.lexeme())
			return s.makeToken(kind)
		case isDigit(s.ch):
			for isDigit(s.ch) {
				s.advance()
			}
			return s.makeToken(token.Num)
		default:
			goto scanError
		}
	}

scanError:
	fmt.Fprintf(os.Stderr, "%s:%d: unexpected character %c\n", s.file, s.line, s.ch)
	os.Exit(1)
	panic("i just wanna exit :(")
}

func isLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isAlpha(ch byte) bool {
	return isLetter(ch) || isDigit(ch)
}
