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

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

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

	// Prefix parsing functions
	par.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	par.registerPrefix(token.IDENT, par.parseIdentifier)
	par.registerPrefix(token.INT, par.parseIntegerLiteral)
	par.registerPrefix(token.TRUE, par.parseBoolean)
	par.registerPrefix(token.FALSE, par.parseBoolean)
	par.registerPrefix(token.BANG, par.parsePrefixExpression)
	par.registerPrefix(token.MINUS, par.parsePrefixExpression)
	par.registerPrefix(token.LPAREN, par.parseGroupedExpression)
	par.registerPrefix(token.IF, par.parseIfExpression)
	par.registerPrefix(token.FUNCTION, par.parseFunctionLiteral)

	// Infix parsing functions
	par.infixParseFns = make(map[token.TokenType]infixParseFn)
	par.registerInfix(token.PLUS, par.parseInfixExpression)
	par.registerInfix(token.MINUS, par.parseInfixExpression)
	par.registerInfix(token.SLASH, par.parseInfixExpression)
	par.registerInfix(token.ASTERISK, par.parseInfixExpression)
	par.registerInfix(token.EQ, par.parseInfixExpression)
	par.registerInfix(token.NOT_EQ, par.parseInfixExpression)
	par.registerInfix(token.LT, par.parseInfixExpression)
	par.registerInfix(token.GT, par.parseInfixExpression)
	par.registerInfix(token.LPAREN, par.parseCallExpression)

	// Read twice to set both curToken and peekToken
	par.advanceTokens()
	par.advanceTokens()

	return par
}

func (par *Parser) Errors() []string {
	return par.errors
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

func (par *Parser) advanceTokens() {
	par.curToken = par.peekToken
	par.peekToken = par.lex.NextToken()
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

	par.advanceTokens()

	stmt.Value = par.parseExpression(LOWEST)

	for !par.curTokenIs(token.SEMICOLON) {
		par.advanceTokens()
	}

	return stmt
}

func (par *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: par.curToken}

	par.advanceTokens()

	stmt.Value = par.parseExpression(LOWEST)

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

	for !par.peekTokenIs(token.SEMICOLON) && precedence < par.peekPrecedence() {
		infix := par.infixParseFns[par.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		par.advanceTokens()

		leftExp = infix(leftExp)
	}

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

func (par *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: par.curToken, Value: par.curTokenIs(token.TRUE)}
}

func (par *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{Token: par.curToken, Operator: par.curToken.Literal}

	par.advanceTokens()
	expression.Right = par.parseExpression(PREFIX)

	return expression
}

func (par *Parser) parseInfixExpression(leftToken ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    par.curToken,
		Operator: par.curToken.Literal,
		Left:     leftToken,
	}

	precedence := par.curPrecedence()
	par.advanceTokens()
	expression.Right = par.parseExpression(precedence)

	return expression
}

func (par *Parser) parseCallExpression(functionIdentifier ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: par.curToken, Function: functionIdentifier}
	exp.Arguments = par.parseCallArguments()
	return exp
}

func (par *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if par.peekTokenIs(token.RPAREN) {
		par.advanceTokens()
		return nil
	}

	par.advanceTokens()
	args = append(args, par.parseExpression(LOWEST))

	for par.peekTokenIs(token.COMMA) {
		par.advanceTokens()
		par.advanceTokens()
		args = append(args, par.parseExpression(LOWEST))
	}

	if !par.peekAssertAdvance(token.RPAREN) {
		return nil
	}

	return args
}

func (par *Parser) parseGroupedExpression() ast.Expression {
	par.advanceTokens()

	exp := par.parseExpression(LOWEST)
	if !par.peekAssertAdvance(token.RPAREN) {
		return nil
	}
	return exp
}

func (par *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: par.curToken}

	if !par.peekAssertAdvance(token.LPAREN) {
		return nil
	}

	par.advanceTokens()
	expression.Condition = par.parseExpression(LOWEST)

	if !par.peekAssertAdvance(token.RPAREN) {
		return nil
	}

	if !par.peekAssertAdvance(token.LBRACE) {
		return nil
	}

	expression.Consequence = par.parseBlockStatement()

	if par.peekTokenIs(token.ELSE) {
		par.advanceTokens()
		if !par.peekAssertAdvance(token.LBRACE) {
			return nil
		}

		expression.Alternative = par.parseBlockStatement()
	}

	return expression
}

func (par *Parser) parseBlockStatement() *ast.BlockStatement {
	block := ast.BlockStatement{Token: par.curToken}
	block.Statements = []ast.Statement{}

	par.advanceTokens()

	for !par.curTokenIs(token.RBRACE) && !par.curTokenIs(token.EOF) {
		stmt := par.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		par.advanceTokens()
	}

	return &block
}

func (par *Parser) parseFunctionLiteral() ast.Expression {
	fl := &ast.FunctionLiteral{Token: par.curToken}

	if !par.peekAssertAdvance(token.LPAREN) {
		return nil
	}

	fl.Parameters = par.parseFunctionParameters()

	if !par.peekAssertAdvance(token.LBRACE) {
		return nil
	}

	fl.Body = par.parseBlockStatement()

	return fl
}

func (par *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// Empty set of function parameters
	if par.peekTokenIs(token.RPAREN) {
		par.advanceTokens()
		return identifiers
	}

	par.advanceTokens()

	ident := &ast.Identifier{Token: par.curToken, Value: par.curToken.Literal}
	identifiers = append(identifiers, ident)

	for par.peekTokenIs(token.COMMA) {
		par.advanceTokens()
		par.advanceTokens()
		ident := &ast.Identifier{Token: par.curToken, Value: par.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !par.peekAssertAdvance(token.RPAREN) {
		return nil
	}

	return identifiers
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

func (par *Parser) curPrecedence() int {
	if precedence, ok := precedences[par.curToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (par *Parser) peekPrecedence() int {
	if precedence, ok := precedences[par.peekToken.Type]; ok {
		return precedence
	}
	return LOWEST
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
