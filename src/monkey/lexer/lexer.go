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

// Get the next token in the Lexer
func (lex *Lexer) NextToken() token.Token {
	var tok token.Token

	lex.skipWhiteSpace()

	switch lex.ch {
	case '=':
		if lex.peekChar() == '=' {
			initialCh := lex.ch
			lex.readChar()
			literal := string(initialCh) + string(lex.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, lex.ch)
		}
	case '+':
		tok = newToken(token.PLUS, lex.ch)
	case '-':
		tok = newToken(token.MINUS, lex.ch)
	case '!':
		if lex.peekChar() == '=' {
			initialCh := lex.ch
			lex.readChar()
			literal := string(initialCh) + string(lex.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, lex.ch)
		}
	case '*':
		tok = newToken(token.ASTERISK, lex.ch)
	case '/':
		tok = newToken(token.SLASH, lex.ch)
	case '<':
		tok = newToken(token.LT, lex.ch)
	case '>':
		tok = newToken(token.GT, lex.ch)
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
	default:
		if isLetter(lex.ch) {
			tok.Literal = lex.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(lex.ch) {
			tok.Literal = lex.readNumber()
			tok.Type = token.INT
			return tok
		} else {
			tok = newToken(token.ILLEGAL, lex.ch)
		}
	}

	lex.readChar()
	return tok
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

// Get the next character without advancing position
func (lex *Lexer) peekChar() byte {
	if lex.readPosition >= len(lex.input) {
		return 0
	}
	return lex.input[lex.readPosition]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func (lex *Lexer) readIdentifier() string {
	initialPosition := lex.position
	for isLetter(lex.ch) {
		lex.readChar()
	}
	return lex.input[initialPosition:lex.position]
}

func (lex *Lexer) readNumber() string {
	initialPosition := lex.position
	for isDigit(lex.ch) {
		lex.readChar()
	}
	return lex.input[initialPosition:lex.position]
}

// Advances through the text until EOF or the next non-whitespace character is found
func (lex *Lexer) skipWhiteSpace() {
	for lex.ch == ' ' || lex.ch == '\t' || lex.ch == '\n' || lex.ch == '\r' {
		lex.readChar()
	}
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
