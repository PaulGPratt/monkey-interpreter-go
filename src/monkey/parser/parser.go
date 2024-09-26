package parser

import (
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

type Parser struct {
	lex *lexer.Lexer

	curToken  token.Token
	peekToken token.Token
}

func New(lex *lexer.Lexer) *Parser {
	par := &Parser{lex: lex}

	// Read twice to set both curToken and peekToken
	par.advanceTokens()
	par.advanceTokens()

	return par
}

func (par *Parser) advanceTokens() {
	par.curToken = par.peekToken
	par.peekToken = par.lex.NextToken()
}

func (par *Parser) ParseProgram() *ast.Program {
	return nil
}
