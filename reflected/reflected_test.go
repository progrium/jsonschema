package reflected

import (
	"fmt"
	"testing"

	"github.com/invopop/jsonschema"
)

type Text string

type Bytes []byte

type Subtype struct {
	Value int
}

type Fooer interface {
	Foo()
}

type ReflectedTest struct {
	Field string
	Sub   Subtype
	Foo   Fooer
}

func (mt ReflectedTest) SimpleMethod() {}

func (mt ReflectedTest) SingleReturn() string {
	return ""
}

func (mt *ReflectedTest) MultiReturn() (int, *bool, error) {
	return 0, nil, nil
}

func (mt *ReflectedTest) SimpleArguments(a int, b string, c bool, d Text, e interface{}) {}

func (mt *ReflectedTest) VariadicArguments(a int, b ...string) {}

func (mt *ReflectedTest) ComplexArguments(a []Subtype, b map[string]Subtype, c *Subtype) {}

func (mt *ReflectedTest) ArgsAndReturn(a []*string, b []interface{}) (Bytes, *Subtype) {
	return nil, nil
}

func TestReflected(t *testing.T) {
	rt := &ReflectedTest{}
	r := &jsonschema.Reflector{
		AnnotatePointers:          true,
		AnnotatePackages:          true,
		AnnotateMethods:           true,
		AnnotateNames:             true,
		AllowAdditionalProperties: true,
	}
	schema := r.Reflect(rt)
	tt := TypeOf(schema)
	for i := 0; i < tt.NumMethod(); i++ {
		m := tt.Method(i)
		var args []Type
		for a := 1; a < m.Type.NumIn(); a++ {
			args = append(args, m.Type.In(a))
		}
		var rets []Type
		for a := 0; a < m.Type.NumOut(); a++ {
			rets = append(rets, m.Type.Out(a))
		}
		fmt.Println(m.Name, args, rets)
	}
	fmt.Println(tt.Field(2).Type.NumMethod())
}
