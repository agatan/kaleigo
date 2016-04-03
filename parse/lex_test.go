package parse

import (
	"reflect"
	"testing"
)

func TestLex(t *testing.T) {
	lexer := lex("test", "abc, 123.4;def ( ) if then else if1 for in")
	expected := []token{
		{kind: tokIdentifier, value: "abc"},
		{kind: tokComma, value: ","},
		{kind: tokNumber, value: "123.4"},
		{kind: tokSemi, value: ";"},
		{kind: tokDef, value: "def"},
		{kind: tokLparen, value: "("},
		{kind: tokRparen, value: ")"},
		{kind: tokIf, value: "if"},
		{kind: tokThen, value: "then"},
		{kind: tokElse, value: "else"},
		{kind: tokIdentifier, value: "if1"},
		{kind: tokFor, value: "for"},
		{kind: tokIn, value: "in"},
		{kind: tokEOF, value: ""},
	}

	for _, e := range expected {
		actual := lexer.nextToken()
		if !reflect.DeepEqual(e, actual) {
			t.Errorf("lex error: expected %#v, actual %#v", e, actual)
		}
	}
}

func TestLexError(t *testing.T) {
	lexer := lex("test", "123.4asd")
	actual := lexer.nextToken()
	if actual.kind != tokError || actual.value != `bad number syntax: "123.4a"` {
		t.Errorf("bad number syntax is not detected: result: %#v", actual)
	}
}
