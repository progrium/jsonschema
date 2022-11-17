package reflected

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/progrium/jsonschema"
)

type Method struct {
	Name    string
	PkgPath string
	Type    Type
}

type StructField struct {
	Name    string
	PkgPath string
	Type    Type
}

type Type struct {
	schema *jsonschema.Schema
	defs   jsonschema.Definitions
	method *jsonschema.Method
}

func resolve(s *jsonschema.Schema, defs jsonschema.Definitions) *jsonschema.Schema {
	if s.Ref == "" || defs == nil {
		return s
	}
	if !strings.HasPrefix(s.Ref, "#/$defs/") {
		return s
	}
	def := defs[strings.TrimPrefix(s.Ref, "#/$defs/")]
	if !s.Pointer {
		return def
	}
	return ptrFrom(def)
}

func ptrFrom(s *jsonschema.Schema) *jsonschema.Schema {
	if s.Pointer {
		return s
	}
	ss := *s
	ss.Pointer = true
	return &ss
}

func indirect(s *jsonschema.Schema) *jsonschema.Schema {
	if !s.Pointer {
		return s
	}
	ss := *s
	ss.Pointer = false
	return &ss
}

func typeFrom(t Type, s *jsonschema.Schema) Type {
	tt := t
	tt.method = nil
	tt.schema = resolve(s, t.defs)
	return tt
}

func TypeOf(s *jsonschema.Schema) Type {
	return Type{schema: resolve(s, s.Definitions), defs: s.Definitions}
}

func (t Type) Kind() reflect.Kind {
	if t.schema.Pointer {
		return reflect.Pointer
	}
	if t.method != nil {
		return reflect.Func
	}
	switch t.schema.Type {
	case "integer":
		return reflect.Int
	case "numeric":
		return reflect.Float64
	case "string":
		if t.schema.ContentEncoding == "base64" {
			return reflect.Slice // of bytes
		}
		return reflect.String
	case "boolean":
		return reflect.Bool
	case "array":
		return reflect.Slice
	case "byte":
		return reflect.Uint8
	case "object":
		if t.schema.PatternProperties != nil {
			return reflect.Map
		}
		return reflect.Struct
	default:
		return reflect.Interface
	}
}

func (t Type) Name() string {
	return t.schema.Name
}

func (t Type) PkgPath() string {
	return t.schema.Package
}

func (t Type) IsVariadic() bool {
	if t.method == nil {
		return false
	}
	return t.method.Variadic
}

func (t Type) NumMethod() int {
	return len(t.schema.Methods.Keys())
}

func (t Type) NumIn() int {
	if t.method.In == nil {
		return 1
	}
	return len(t.method.In) + 1
}

func (t Type) NumOut() int {
	if t.method.Out == nil {
		return 0
	}
	return len(t.method.Out)
}

func (t Type) NumField() int {
	if t.schema.Properties == nil {
		return 0
	}
	return len(t.schema.Properties.Keys())
}

func (t Type) Method(i int) Method {
	name := t.schema.Methods.Keys()[i]
	m, _ := t.MethodByName(name)
	return m
}

func (t Type) MethodByName(name string) (Method, bool) {
	s, ok := t.schema.Methods.Get(name)
	if ok {
		tt := typeFrom(t, indirect(t.schema))
		m := s.(jsonschema.Method)
		tt.method = &m
		return Method{
			Name:    name,
			PkgPath: t.schema.Package,
			Type:    tt,
		}, true
	}
	return Method{}, false
}

func (t Type) Field(i int) StructField {
	name := t.schema.Properties.Keys()[i]
	f, _ := t.FieldByName(name)
	return f
}

func (t Type) FieldByName(name string) (StructField, bool) {
	s, ok := t.schema.Properties.Get(name)
	if ok {
		return StructField{
			Name:    name,
			PkgPath: t.schema.Package,
			Type:    typeFrom(t, s.(*jsonschema.Schema)),
		}, true
	}
	return StructField{}, false
}

func (t Type) In(i int) Type {
	if i == 0 {
		return t
	}
	return typeFrom(t, t.method.In[i-1])
}

func (t Type) Out(i int) Type {
	return typeFrom(t, t.method.Out[i])
}

func (t Type) Key() Type {
	return Type{schema: &jsonschema.Schema{Type: "string"}}
}

func (t Type) Elem() Type {
	switch t.Kind() {
	case reflect.Pointer:
		return typeFrom(t, indirect(t.schema))
	case reflect.Map:
		return typeFrom(t, t.schema.PatternProperties[".*"])
	case reflect.Slice:
		if t.schema.ContentEncoding == "base64" {
			return Type{schema: &jsonschema.Schema{Type: "byte"}}
		}
		return typeFrom(t, t.schema.Items)
	default:
		return Type{}
	}
}

func (t Type) String() string {
	if t.Kind() == reflect.Func {
		return "func"
	}
	if t.Kind() == reflect.Slice && t.schema.Name == "" {
		return "[]" + t.Elem().String()
	}
	if t.Kind() == reflect.Map && t.schema.Name == "" {
		return fmt.Sprintf("map[%s]%s", t.Key().String(), t.Elem().String())
	}
	name := t.Kind().String()
	if t.schema.Name != "" {
		name = t.schema.Name
	}
	if t.schema.Pointer {
		return "*" + t.Elem().String()
	}
	if t.Kind() == reflect.Interface && t.schema.Name == "" {
		return name + "{}"
	}
	return name
}
