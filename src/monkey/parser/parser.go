package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // less than (<) or greater than (>)
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	lex    *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(lex *lexer.Lexer) *Parser {
	par := &Parser{
		lex:    lex,
		errors: []string{},
	}

	par.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	par.registerPrefix(token.IDENT, par.parseIdentifier)
	par.registerPrefix(token.INT, par.parseIntegerLiteral)
	par.registerPrefix(token.BANG, par.parsePrefixExpression)
	par.registerPrefix(token.MINUS, par.parsePrefixExpression)

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
		return par.parseExpressionStatement()
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

func (par *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: par.curToken}

	stmt.Expression = par.parseExpression(LOWEST)

	if par.peekTokenIs(token.SEMICOLON) {
		par.advanceTokens()
	}

	return stmt
}

func (par *Parser) parseExpression(precedence int) ast.Expression {
	prefix := par.prefixParseFns[par.curToken.Type]
	if prefix == nil {
		par.noPrefixParseFnError(par.curToken.Type)
		return nil
	}
	leftExp := prefix()

	return leftExp
}

func (par *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: par.curToken, Value: par.curToken.Literal}
}

func (par *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: par.curToken}

	value, err := strconv.ParseInt(par.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", par.curToken.Literal)
		par.errors = append(par.errors, msg)
	}

	lit.Value = value
	return lit
}

func (par *Parser) parsePrefixExpression() ast.Expression {
	exp := &ast.PrefixExpression{Token: par.curToken, Operator: par.curToken.Literal}

	par.advanceTokens()
	exp.Right = par.parseExpression(PREFIX)

	return exp
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

func (par *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	par.prefixParseFns[tokenType] = fn
}

func (par *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	par.infixParseFns[tokenType] = fn
}

func (par *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	par.errors = append(par.errors, msg)
}
