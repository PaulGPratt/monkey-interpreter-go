package lexer

import "monkey/token"

type Lexer struct {
	input        string
	position     int  // current char position in input (points to ch)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	lex := &Lexer{input: input}
	lex.readChar()
	return lex
}

// Get the next character and advance read position by 1
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII for NUL character
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

// Get the next token in the Lexer
func (lex *Lexer) NextToken() token.Token {
	var tok token.Token

	switch lex.ch {
	case '=':
		tok = newToken(token.ASSIGN, lex.ch)
	case '+':
		tok = newToken(token.PLUS, lex.ch)
	case ',':
		tok = newToken(token.COMMA, lex.ch)
	case ';':
		tok = newToken(token.SEMICOLON, lex.ch)
	case '(':
		tok = newToken(token.LPAREN, lex.ch)
	case ')':
		tok = newToken(token.RPAREN, lex.ch)
	case '{':
		tok = newToken(token.LBRACE, lex.ch)
	case '}':
		tok = newToken(token.RBRACE, lex.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	}

	lex.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
