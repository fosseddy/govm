package scanner

import (
	"fmt"
	"os"
	"strconv"
	"asm/token"
)

type scanner struct {
	file string
	src []byte
	line int
	start int
	cur int
	ch byte
}

func Scan(file string) []token.Token {
	src, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	s := scanner{
		file: file,
		src: src,
		line: 1,
	}


	if len(src) > 0 {
		s.ch = src[0]
	}

	toks := make([]token.Token, 0, 256)

	for {
		tok := s.scanToken()
		toks = append(toks, tok)
		if tok.Kind == token.EOF {
			break
		}
	}

	return toks
}

func (s *scanner) hasSrc() bool {
	return s.cur < len(s.src)
}

func (s *scanner) advance() {
	s.cur++

	if !s.hasSrc() {
		return
	}

	s.ch = s.src[s.cur]
}

func (s *scanner) next(ch byte) bool {
	next := s.cur + 1
	return next < len(s.src) && s.src[next] == ch
}

func (s *scanner) makeToken(kind token.Kind) token.Token {
	tok := token.Token{
		Kind: kind,
		Lex: s.lexeme(),
		Pos: token.Position{s.file, s.line},
	}

	switch tok.Kind {
	case token.Char:
		tok.Value = int(tok.Lex[1])
	case token.Num:
		val, err := strconv.Atoi(tok.Lex)
		if err != nil {
			// TODO(art): better error message
			s.reportError("%s\n", err)
		}
		tok.Value = val
	}

	return tok
}

func (s *scanner) lexeme() string {
	return string(s.src[s.start:s.cur])
}

func (s *scanner) reportError(fstr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s:%d: ", s.file, s.line)
	fmt.Fprintf(os.Stderr, fstr, args...)
	os.Exit(1)
}

func (s *scanner) scanToken() token.Token {
scanAgain:
	s.start = s.cur

	if !s.hasSrc() {
		return s.makeToken(token.EOF)
	}

	switch s.ch {
	case ' ', '\t', '\r':
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

	case '\n':
		s.advance()
		t := s.makeToken(token.LF)
		s.line++
		return t

	case '\'':
		s.advance()
		for s.ch != '\'' && s.ch != '\n' && s.hasSrc() {
			s.advance()
		}

		if s.ch == '\n' {
			s.reportError("unterminated character literal\n")
		}

		ch := s.src[s.start + 1:s.cur]
		if len(ch) > 1 {
			s.reportError("expected single character\n")
		}

		s.advance()
		return s.makeToken(token.Char)
	case '"':
		s.advance()
		for s.ch != '"' && s.ch != '\n' && s.hasSrc() {
			s.advance()
		}

		if s.ch == '\n' {
			s.reportError("unterminated string literal\n")
		}

		str := s.src[s.start + 1:s.cur]
		if len(str) == 0 {
			s.reportError("empty string literal\n")
		}

		s.advance()
		return s.makeToken(token.Str)
	
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

	goto scanAgain

scanError:
	s.reportError("unexpected character %c\n", s.ch)
	panic("reportError calls os.Exit(1)")
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
