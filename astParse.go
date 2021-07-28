package astParse

// ast 针对xss,jsonP 只需要用到 StringLiteral Identifier,来查找对应的关键字即可

import (
	"fmt"
	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
	"github.com/jvatic/goja-babel"
	"reflect"

	"strings"
)

//WhitespaceChars := " \f\n\r\t\v\u00a0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a\u2028\u2029\u202f\u205f\u3000\ufeff"

var (
	TypeScriptIdentifier         = "TypeScriptIdentifier"
	TypeScriptVariableExpression = "TypeScriptVariableExpression"
	TypeScriptStringLiteral      = "TypeScriptStringLiteral"
	TypeScriptNumberLiteral      = "TypeScriptNumberLiteral"

)

type JSToken struct {
	Content string
	Type    string
	Index   int
}

type JSParse struct {
	tokenizers []*JSToken
	expressType reflect.Type
}

func (p *JSParse) Init() {
	err := babel.Init(1)
	fmt.Println(err)
}

func New() * JSParse{
	return &JSParse{
		tokenizers: []*JSToken{},
		expressType :reflect.TypeOf(new(ast.Expression)).Elem(),
	}
}

// TransFormString 转换es6到es5
func (p *JSParse) TransFormString(content string) (string, error) {
	p.Init()
	res, err := babel.TransformString(content, map[string]interface{}{"presets": []string{"env"}})
	if err != nil {
		return "", err
	}
	res = strings.Replace(res, "\"use strict\";", "", 1)
	return res, nil
}

func (p *JSParse) ParseProgram(content string) (*ast.Program, error) {
	program, err := parser.ParseFile(nil, "", content, 0)
	if err != nil {
		return nil, err
	}
	return program, nil
}

func (p *JSParse)GetToken()[]*JSToken{
	return p.tokenizers
}


func (p *JSParse) analyseNode(node interface{}) {
	if id, ok := node.(*ast.Identifier); ok && id != nil {
		t := &JSToken{
			Type:    TypeScriptIdentifier,
			Index:   int(id.Idx0()),
			Content: string(id.Name),
		}
		p.tokenizers = append(p.tokenizers, t)
	} else if v, ok := node.(*ast.VariableExpression); ok && v != nil {
		t := &JSToken{
			Type:  TypeScriptVariableExpression,
			Index: int(v.Idx0()),
			Content: string(v.Name),
		}
		p.tokenizers = append(p.tokenizers, t)
	} else if s, ok := node.(*ast.StringLiteral); ok && s != nil {
		t := &JSToken{
			Type:    TypeScriptStringLiteral,
			Content: s.Literal,
			Index:   int(s.Idx0()),
		}
		p.tokenizers = append(p.tokenizers, t)
	} else if n,ok := node.(*ast.NumberLiteral);ok && n != nil{
		t := &JSToken{
			Type:  TypeScriptNumberLiteral,
			Index: int(n.Idx0()),
			Content: n.Literal,
		}
		p.tokenizers = append(p.tokenizers, t)
	}

}

func (p *JSParse) ForEach(statement interface{}){
	getValue, flag := detectType(statement)

	switch flag{
	case "":
		return
	case "slice":
		l := getValue.Len()
		for i:=0;i<l;i++{
			value := getValue.Index(i)
			flagExpress := value.Type().Implements(p.expressType)
			if flagExpress{
				p.analyseNode(value.Interface())
			}
			p.ForEach(value)

		}
	case "struct":
		flagExpress := getValue.Type().Implements(p.expressType)
		if flagExpress{
			p.analyseNode(getValue.Interface())
		}else{
			for i := 0; i < getValue.Type().NumField(); i++ {
				value := getValue.Field(i)
				p.ForEach(value)
			}
		}
	case "interface":
		flagExpress := getValue.Type().Implements(p.expressType)
		if flagExpress{
			p.analyseNode(getValue.Interface())
		}
		p.ForEach(getValue.Interface())

	}


}

func detectType(input interface{}) (reflect.Value,string){
	var getValue reflect.Value
	if v, ok:= input.(reflect.Value);ok{
		getValue = v
	}else{
		getValue = reflect.ValueOf(input)
	}
	if getValue.Kind() == reflect.Ptr{
		getValue = getValue.Elem()
	}
	if getValue.Kind() == reflect.Slice{
		return getValue,"slice"
	}
	if getValue.Kind() == reflect.Struct{
		return getValue,"struct"
	}
	if getValue.Kind() ==  reflect.Interface{
		return getValue, "interface"
	}
	return reflect.Value{}, ""
}
