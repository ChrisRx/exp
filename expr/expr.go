package expr

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"

	"go.chrisrx.dev/x/slices"
	"go.chrisrx.dev/x/strings"
)

// TODO(ChrisRx): pass in variables map
// TODO(ChrisRx): expose hook function(s)
// TODO(ChrisRx): make struct with default builtins/funcs that can be modified
// with constructor options

type Option func(*Expr)

func Env(env map[string]reflect.Value) Option {
	return func(e *Expr) {
		e.env = env
	}
}

type Expr struct {
	env map[string]reflect.Value
}

func Eval(s string, opts ...Option) (reflect.Value, error) {
	s = strings.ReplaceAll(s, `'`, `"`)
	expr, err := parser.ParseExpr(s)
	if err != nil {
		return reflect.Value{}, err
	}
	e := new(Expr)
	for _, opt := range opts {
		opt(e)
	}
	// TODO(ChrisRx): check if CanInterface()?
	return e.eval(expr)
}

func upper(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	return strings.Join(
		slices.Map(
			strings.Split(string(unicode.ToUpper(r))+s[size:], "_"),
			strings.Title,
		), "")
}

var errUndefined = errors.New("undefined")

func (e *Expr) eval(expr ast.Expr) (reflect.Value, error) {
	switch expr := expr.(type) {
	case *ast.ParenExpr:
		return e.eval(expr.X)
	case *ast.StarExpr:
		return e.eval(expr.X)
	case *ast.CallExpr:
		return e.evalCallExpr(expr)
	case *ast.SelectorExpr:
		return e.evalSelectorExpr(expr)
	case *ast.Ident:
		if v, ok := e.env[expr.Name]; ok {
			return v, nil
		}
		if v, ok := builtins[expr.Name]; ok {
			return v, nil
		}
		if _, ok := packages()[expr.Name]; ok {
			return reflect.ValueOf(expr.Name), nil
		}
		return reflect.Value{}, fmt.Errorf("%w: %v", errUndefined, expr.Name)
	case *ast.BasicLit:
		return e.evalBasicLit(expr)
	case *ast.CompositeLit:
		return e.evalCompositeLit(expr)
	case *ast.UnaryExpr:
		return e.evalUnaryExpr(expr)
	case *ast.BinaryExpr:
		return e.evalBinaryExpr(expr)
	case *ast.IndexExpr:
		return e.evalIndexExpr(expr)
	case *ast.SliceExpr:
		return e.evalSliceExpr(expr)
	}
	return reflect.Value{}, fmt.Errorf("unsupported ast.Expr: %T", expr)
}

func (e *Expr) evalBasicLit(expr *ast.BasicLit) (reflect.Value, error) {
	switch expr.Kind {
	case token.INT:
		i, err := strconv.ParseInt(expr.Value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(i), nil
	case token.FLOAT:
		f, err := strconv.ParseFloat(expr.Value, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(f), nil
	case token.STRING, token.CHAR:
		return reflect.ValueOf(expr.Value[1 : len(expr.Value)-1]), nil
	}
	return reflect.Value{}, fmt.Errorf("unsupported *ast.BasicLit: %T", expr.Kind)
}

func (e *Expr) evalBinaryExpr(expr *ast.BinaryExpr) (reflect.Value, error) {
	lhs, err := e.eval(expr.X)
	if err != nil {
		return reflect.Value{}, err
	}
	rhs, err := e.eval(expr.Y)
	if err != nil {
		return reflect.Value{}, err
	}

	switch lhs := lhs.Interface().(type) {
	case time.Duration:
		switch rhs := rhs.Interface().(type) {
		case time.Time:
			switch expr.Op {
			case token.ADD:
				return reflect.ValueOf(rhs.Add(lhs)), nil
			case token.SUB:
				return reflect.ValueOf(rhs.Add(-lhs)), nil
			}
		case *time.Time:
			switch expr.Op {
			case token.ADD:
				return reflect.ValueOf(rhs.Add(lhs)), nil
			case token.SUB:
				return reflect.ValueOf(rhs.Add(-lhs)), nil
			}
		}
	case time.Time:
		switch rhs := rhs.Interface().(type) {
		case time.Duration:
			switch expr.Op {
			case token.ADD:
				return reflect.ValueOf(lhs.Add(rhs)), nil
			case token.SUB:
				return reflect.ValueOf(lhs.Add(-rhs)), nil
			}
		}
	case *time.Time:
		switch rhs := rhs.Interface().(type) {
		case time.Duration:
			switch expr.Op {
			case token.ADD:
				return reflect.ValueOf(lhs.Add(rhs)), nil
			case token.SUB:
				return reflect.ValueOf(lhs.Add(-rhs)), nil
			}
		}
	}

	if lhs.Type() != rhs.Type() {
		if !lhs.Type().ConvertibleTo(rhs.Type()) {
			return reflect.Value{}, fmt.Errorf("invalid operation: operator %s on incompatible types: %v <> %v", expr.Op, lhs.Type(), rhs.Type())
		}
	}

	switch {
	case lhs.CanFloat() || rhs.CanFloat():
		floatType := reflect.TypeOf(float64(0))
		lhs, rhs = lhs.Convert(floatType), rhs.Convert(floatType)
	case lhs.CanInt() && lhs.CanUint():
		if lhs.Type().Bits() > rhs.Type().Bits() {
			rhs = rhs.Convert(lhs.Type())
		}
		lhs = lhs.Convert(rhs.Type())
	}

	switch lhs.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch expr.Op {
		case token.ADD:
			return reflect.ValueOf(lhs.Int() + rhs.Int()), nil
		case token.SUB:
			return reflect.ValueOf(lhs.Int() - rhs.Int()), nil
		case token.MUL:
			return reflect.ValueOf(lhs.Int() * rhs.Int()), nil
		case token.QUO:
			return reflect.ValueOf(lhs.Int() / rhs.Int()), nil
		case token.REM:
			return reflect.ValueOf(lhs.Int() % rhs.Int()), nil
		case token.AND:
			return reflect.ValueOf(lhs.Int() & rhs.Int()), nil
		case token.OR:
			return reflect.ValueOf(lhs.Int() | rhs.Int()), nil
		case token.XOR:
			return reflect.ValueOf(lhs.Int() ^ rhs.Int()), nil
		case token.EQL:
			return reflect.ValueOf(lhs.Int() == rhs.Int()), nil
		case token.LSS:
			return reflect.ValueOf(lhs.Int() < rhs.Int()), nil
		case token.GTR:
			return reflect.ValueOf(lhs.Int() > rhs.Int()), nil
		case token.NEQ:
			return reflect.ValueOf(lhs.Int() != rhs.Int()), nil
		case token.LEQ:
			return reflect.ValueOf(lhs.Int() <= rhs.Int()), nil
		case token.GEQ:
			return reflect.ValueOf(lhs.Int() >= rhs.Int()), nil
		case token.SHL:
			return reflect.ValueOf(lhs.Int() << rhs.Int()), nil
		case token.SHR:
			return reflect.ValueOf(lhs.Int() >> rhs.Int()), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch expr.Op {
		case token.ADD:
			return reflect.ValueOf(lhs.Uint() + rhs.Uint()), nil
		case token.SUB:
			return reflect.ValueOf(lhs.Uint() - rhs.Uint()), nil
		case token.MUL:
			return reflect.ValueOf(lhs.Uint() * rhs.Uint()), nil
		case token.QUO:
			return reflect.ValueOf(lhs.Uint() / rhs.Uint()), nil
		case token.REM:
			return reflect.ValueOf(lhs.Uint() % rhs.Uint()), nil
		case token.AND:
			return reflect.ValueOf(lhs.Uint() & rhs.Uint()), nil
		case token.OR:
			return reflect.ValueOf(lhs.Uint() | rhs.Uint()), nil
		case token.XOR:
			return reflect.ValueOf(lhs.Uint() ^ rhs.Uint()), nil
		case token.EQL:
			return reflect.ValueOf(lhs.Uint() == rhs.Uint()), nil
		case token.LSS:
			return reflect.ValueOf(lhs.Uint() < rhs.Uint()), nil
		case token.GTR:
			return reflect.ValueOf(lhs.Uint() > rhs.Uint()), nil
		case token.NEQ:
			return reflect.ValueOf(lhs.Uint() != rhs.Uint()), nil
		case token.LEQ:
			return reflect.ValueOf(lhs.Uint() <= rhs.Uint()), nil
		case token.GEQ:
			return reflect.ValueOf(lhs.Uint() >= rhs.Uint()), nil
		case token.SHL:
			return reflect.ValueOf(lhs.Uint() << rhs.Uint()), nil
		case token.SHR:
			return reflect.ValueOf(lhs.Uint() >> rhs.Uint()), nil
		}
	case reflect.Float32, reflect.Float64:
		switch expr.Op {
		case token.ADD:
			return reflect.ValueOf(lhs.Float() + rhs.Float()), nil
		case token.SUB:
			return reflect.ValueOf(lhs.Float() - rhs.Float()), nil
		case token.MUL:
			return reflect.ValueOf(lhs.Float() * rhs.Float()), nil
		case token.QUO:
			return reflect.ValueOf(lhs.Float() / rhs.Float()), nil
		case token.EQL:
			return reflect.ValueOf(lhs.Float() == rhs.Float()), nil
		case token.LSS:
			return reflect.ValueOf(lhs.Float() < rhs.Float()), nil
		case token.GTR:
			return reflect.ValueOf(lhs.Float() > rhs.Float()), nil
		case token.NEQ:
			return reflect.ValueOf(lhs.Float() != rhs.Float()), nil
		case token.LEQ:
			return reflect.ValueOf(lhs.Float() <= rhs.Float()), nil
		case token.GEQ:
			return reflect.ValueOf(lhs.Float() >= rhs.Float()), nil
		}
	case reflect.String:
		switch expr.Op {
		case token.ADD:
			return reflect.ValueOf(lhs.String() + rhs.String()), nil
		case token.SUB:
			return reflect.ValueOf(lhs.String() + rhs.String()), nil
		case token.EQL:
			return reflect.ValueOf(lhs.Equal(rhs)), nil
		case token.LSS:
			return reflect.ValueOf(lhs.String() < rhs.String()), nil
		case token.GTR:
			return reflect.ValueOf(lhs.String() > rhs.String()), nil
		case token.NEQ:
			return reflect.ValueOf(!lhs.Equal(rhs)), nil
		case token.LEQ:
			return reflect.ValueOf(lhs.String() <= rhs.String()), nil
		case token.GEQ:
			return reflect.ValueOf(lhs.String() >= rhs.String()), nil
		}
	case reflect.Bool:
		switch expr.Op {
		case token.EQL:
			return reflect.ValueOf(lhs.Bool() == rhs.Bool()), nil
		case token.NEQ:
			return reflect.ValueOf(lhs.Bool() != rhs.Bool()), nil
		case token.LAND:
			return reflect.ValueOf(lhs.Bool() && rhs.Bool()), nil
		case token.LOR:
			return reflect.ValueOf(lhs.Bool() || rhs.Bool()), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("operator %s not defined for %v", expr.Op, lhs.Type())
}

func (e *Expr) evalCallExpr(expr *ast.CallExpr) (reflect.Value, error) {
	fn, err := e.eval(expr.Fun)
	if err != nil {
		return reflect.Value{}, err
	}
	if fn.Kind() != reflect.Func {
		rv := reflect.New(fn.Type()).Elem()
		if len(expr.Args) == 0 {
			return rv, nil
		}
		v, err := e.eval(expr.Args[0])
		if err != nil {
			return v, err
		}
		rv.Set(v.Convert(fn.Type()))
		return rv, nil
		// return reflect.Value{}, fmt.Errorf("unsupported *ast.CallExpr, must be function: (%v)(%v)", fn.Type(), fn)
	}
	var args []reflect.Value
	if fn.Type().IsVariadic() && len(expr.Args) >= fn.Type().NumIn() {
		vt := fn.Type().In(max(fn.Type().NumIn()-1, 0)).Elem()
		for _, arg := range expr.Args {
			v, err := e.eval(arg)
			if err != nil {
				return v, err
			}
			switch {
			case v.Type().AssignableTo(vt):
				args = append(args, v)
			case v.CanConvert(vt):
				args = append(args, v.Convert(vt))
			default:
				return reflect.Value{}, fmt.Errorf("idk")
			}
		}
		results := fn.Call(args)
		return results[0], nil
	}
	if len(expr.Args) != fn.Type().NumIn() && fn.Type().NumIn()-1 != len(expr.Args) {
		name, ok := expr.Fun.(*ast.Ident)
		if !ok {
			name = &ast.Ident{Name: ""}
		}
		return reflect.Value{}, fmt.Errorf("func %s() takes %d args, received %d", name, fn.Type().NumIn(), len(expr.Args))
	}
	for i, arg := range expr.Args {
		v, err := e.eval(arg)
		if err != nil {
			return v, err
		}
		if !v.CanConvert(fn.Type().In(i)) {
			return reflect.Value{}, fmt.Errorf("cannot convert: %v -> %v", v.Type(), fn.Type().In(i))
		}
		args = append(args, v.Convert(fn.Type().In(i)))
	}
	results := fn.Call(args)
	return results[0], nil
}

func (e *Expr) evalCompositeLit(expr *ast.CompositeLit) (reflect.Value, error) {
	t, err := e.eval(expr.Type)
	if err != nil {
		return reflect.Value{}, err
	}
	switch t.Kind() {
	case reflect.Struct:
		rv := reflect.New(t.Type()).Elem()
		for _, elem := range expr.Elts {
			switch elem := elem.(type) {
			case *ast.KeyValueExpr:
				switch key := elem.Key.(type) {
				case *ast.Ident:
					field := rv.FieldByName(key.Name)
					value, err := e.eval(elem.Value)
					if err != nil {
						return reflect.Value{}, err
					}
					field.Set(value.Convert(field.Type()))
				default:
					panic(fmt.Errorf("%T", key))
				}
			}
		}
		return rv, nil
	default:
		return t, nil
	}
}

func (e *Expr) evalSelectorExpr(expr *ast.SelectorExpr) (reflect.Value, error) {
	x, err := e.eval(expr.X)
	if err != nil {
		return reflect.Value{}, err
	}
	x = reflect.Indirect(x)
	switch x.Kind() {
	case reflect.String:
		if pkg, ok := packages()[x.String()]; ok {
			if fn, ok := pkg[expr.Sel.Name]; ok {
				return fn, nil
			}
			if fn, ok := pkg[upper(expr.Sel.Name)]; ok {
				return fn, nil
			}
			return reflect.Value{}, fmt.Errorf("cannot find object in package: %s.%s", x, expr.Sel.Name)
		}
		return reflect.Value{}, fmt.Errorf("cannot find package: %s", x)
	case reflect.Struct:
		method := x.MethodByName(expr.Sel.Name)
		if method.IsValid() {
			return method, nil
		}
		method = x.MethodByName(upper(expr.Sel.Name))
		if method.IsValid() {
			return method, nil
		}
		field := x.FieldByName(expr.Sel.Name)
		if field.IsValid() {
			return field, nil
		}
		field = x.FieldByName(upper(expr.Sel.Name))
		if field.IsValid() {
			return field, nil
		}
		return reflect.Value{}, fmt.Errorf("no field or method %q on %v", expr.Sel.Name, x.Type())
	}
	return reflect.Value{}, fmt.Errorf("unsupported *ast.SelectorExpr: (%v)(%v)", x.Type(), x)
}

func (e *Expr) evalUnaryExpr(expr *ast.UnaryExpr) (reflect.Value, error) {
	v, err := e.eval(expr.X)
	if err != nil {
		return reflect.Value{}, err
	}
	switch expr.Op {
	case token.ADD:
		return reflect.ValueOf(+v.Int()), nil
	case token.SUB:
		return reflect.ValueOf(-v.Int()), nil
	case token.MUL:
		return reflect.ValueOf(v.Pointer()), nil
	case token.AND:
		return reflect.ValueOf(^v.Int()), nil
	case token.NOT:
		// if v.Kind() == reflect.Interface {
		// 	b, ok := v.Interface().(bool)
		// 	if !ok {
		// 		return reflect.Value{}, fmt.Errorf("received invalid type %T for unary operator %T", v.Interface(), expr.Op)
		// 	}
		// 	v = reflect.ValueOf(b)
		// }
		return reflect.ValueOf(!v.Bool()), nil
	case token.XOR:
		return reflect.ValueOf(^v.Int()), nil
	}
	return reflect.Value{}, fmt.Errorf("unsupported *ast.UnaryExpr: %T", expr.Op)
}

func (e *Expr) evalIndexExpr(expr *ast.IndexExpr) (reflect.Value, error) {
	v, err := e.eval(expr.X)
	if err != nil {
		return reflect.Value{}, err
	}
	index, err := e.eval(expr.Index)
	if err != nil {
		return reflect.Value{}, err
	}

	switch index.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		elem := v.Index(int(index.Int()))
		if elem.IsValid() {
			return elem, nil
		}
	}

	return reflect.Value{}, fmt.Errorf("*ast.IndexExpr unsupported")
}

func (e *Expr) evalSliceExpr(expr *ast.SliceExpr) (reflect.Value, error) {
	return reflect.Value{}, fmt.Errorf("*ast.SliceExpr unsupported")
}
