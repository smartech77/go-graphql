package graphql

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/neelance/graphql-go/internal/query"
	"github.com/neelance/graphql-go/internal/schema"
)

type Schema struct {
	*schema.Schema
	resolver reflect.Value
}

func NewSchema(schemaString string, filename string, resolver interface{}) (res *Schema, errRes error) {
	s, err := schema.Parse(schemaString, filename)
	if err != nil {
		return nil, err
	}

	// TODO type check resolver
	return &Schema{
		Schema:   s,
		resolver: reflect.ValueOf(resolver),
	}, nil
}

func (s *Schema) Exec(queryString string) (res []byte, errRes error) {
	q, err := query.Parse(queryString)
	if err != nil {
		return nil, err
	}

	rawRes := exec(s, s.Types["Query"], q, s.resolver)
	return json.Marshal(rawRes)
}

func exec(s *Schema, t schema.Type, sel *query.SelectionSet, resolver reflect.Value) interface{} {
	switch t := t.(type) {
	case *schema.Scalar:
		return resolver.Interface()
	case *schema.Array:
		a := make([]interface{}, resolver.Len())
		for i := range a {
			a[i] = exec(s, t.Elem, sel, resolver.Index(i))
		}
		return a
	case *schema.TypeName:
		return exec(s, s.Types[t.Name], sel, resolver)
	case *schema.Object:
		res := make(map[string]interface{})
		for _, f := range sel.Selections {
			m := findMethod(resolver.Type(), f.Name)
			res[f.Name] = exec(s, t.Fields[f.Name], f.Sel, resolver.Method(m).Call(nil)[0])
		}
		return res
	}
	return nil
}

func findMethod(t reflect.Type, name string) int {
	for i := 0; i < t.NumMethod(); i++ {
		if strings.EqualFold(name, t.Method(i).Name) {
			return i
		}
	}
	return -1
}
