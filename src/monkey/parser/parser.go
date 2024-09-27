package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

type Parser struct {
	lex *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []string
}

func New(lex *lexer.Lexer) *Parser {
	par := &Parser{
		lex:    lex,
		errors: []string{},
	}

	// Read twice to set both curToken and peekToken
	par.advanceTokens()
	par.advanceTokens()

	return par
}

func (par *Parser) Errors() []string {
	return par.errors
}

func (par *Parser) advanceTokens() {
	par.curToken = par.peekToken
	par.peekToken = par.lex.NextToken()
}

func (par *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !par.curTokenIs(token.EOF) {
		stmt := par.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		par.advanceTokens()
	}
	return program
}

func (par *Parser) parseStatement() ast.Statement {
	switch par.curToken.Type {
	case token.LET:
		return par.parseLetStatement()
	case token.RETURN:
		return par.parseReturnStatement()
	default:
		return nil
	}
}

func (par *Parser) parseLetStatement() ast.Statement {
	stmt := &ast.LetStatement{Token: par.curToken}

	if !par.peekAssertAdvance(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: par.curToken, Value: par.curToken.Literal}

	if !par.peekAssertAdvance(token.ASSIGN) {
		return nil
	}

	//TODO: Skipping expressions for now (until I learn how to do it)

	for !par.curTokenIs(token.SEMICOLON) {
		par.advanceTokens()
	}

	return stmt
}

func (par *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: par.curToken}

	par.advanceTokens()

	//TODO: Skipping expressions for now (until I learn how to do it)

	for !par.curTokenIs(token.SEMICOLON) {
		par.advanceTokens()
	}

	return stmt
}

func (par *Parser) curTokenIs(t token.TokenType) bool {
	return par.curToken.Type == t
}

func (par *Parser) peekTokenIs(t token.TokenType) bool {
	return par.peekToken.Type == t
}

func (par *Parser) peekAssertAdvance(t token.TokenType) bool {
	if par.peekTokenIs(t) {
		par.advanceTokens()
		return true
	} else {
		par.appendNextTokenError(t)
		return false
	}
}

func (par *Parser) appendNextTokenError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, but got %s instead", t, par.peekToken.Type)
	par.errors = append(par.errors, msg)
}
