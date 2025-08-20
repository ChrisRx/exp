package expr

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
)

// TODO(ChrisRx): pass in variables map
// TODO(ChrisRx): expose hook function(s)
// TODO(ChrisRx): make struct with default builtins/funcs that can be modified
// with constructor options

func Eval(s string) (reflect.Value, error) {
	expr, err := parser.ParseExpr(s)
	if err != nil {
		return reflect.Value{}, err
	}
	return eval(expr)
}

func eval(expr ast.Expr) (reflect.Value, error) {
	switch expr := expr.(type) {
	case *ast.ParenExpr:
		return eval(expr.X)
	case *ast.StarExpr:
		return eval(expr.X)
	case *ast.CallExpr:
		return evalCallExpr(expr)
	case *ast.SelectorExpr:
		return evalSelectorExpr(expr)
	case *ast.Ident:
		if builtin, ok := builtins[expr.Name]; ok {
			return builtin, nil
		}
		return reflect.ValueOf(expr.Name), nil
	case *ast.BasicLit:
		return evalBasicLit(expr)
	case *ast.CompositeLit:
		return evalCompositeLit(expr)
	case *ast.UnaryExpr:
		return evalUnaryExpr(expr)
	case *ast.BinaryExpr:
		return evalBinaryExpr(expr)
	case *ast.IndexExpr:
		return evalIndexExpr(expr)
	case *ast.SliceExpr:
		return evalSliceExpr(expr)
	}
	return reflect.Value{}, fmt.Errorf("unsupported ast.Expr: %T", expr)
}

func evalBasicLit(expr *ast.BasicLit) (reflect.Value, error) {
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

func evalBinaryExpr(expr *ast.BinaryExpr) (reflect.Value, error) {
	lhs, err := eval(expr.X)
	if err != nil {
		return reflect.Value{}, err
	}
	rhs, err := eval(expr.Y)
	if err != nil {
		return reflect.Value{}, err
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
	case lhs.CanInt() || lhs.CanUint():
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

func evalCallExpr(expr *ast.CallExpr) (reflect.Value, error) {
	fn, err := eval(expr.Fun)
	if err != nil {
		return fn, err
	}
	if fn.Kind() != reflect.Func {
		rv := reflect.New(fn.Type()).Elem()
		if len(expr.Args) == 0 {
			return rv, nil
		}
		v, err := eval(expr.Args[0])
		if err != nil {
			return v, err
		}
		rv.Set(v.Convert(fn.Type()))
		return rv, nil
		// return reflect.Value{}, fmt.Errorf("unsupported *ast.CallExpr, must be function: (%v)(%v)", fn.Type(), fn)
	}
	var args []reflect.Value
	if fn.Type().IsVariadic() {
		vt := fn.Type().In(len(expr.Args) - fn.Type().NumIn()).Elem()
		for _, arg := range expr.Args {
			v, err := eval(arg)
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
	for i, arg := range expr.Args {
		v, err := eval(arg)
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

func evalCompositeLit(expr *ast.CompositeLit) (reflect.Value, error) {
	t, err := eval(expr.Type)
	if err != nil {
		return reflect.Value{}, err
	}
	switch t.Kind() {
	case reflect.Struct:
		rv := reflect.New(t.Type()).Elem()
		for _, elem := range expr.Elts {
			switch elem := elem.(type) {
			case *ast.KeyValueExpr:
				key, err := eval(elem.Key)
				if err != nil {
					return reflect.Value{}, err
				}
				field := rv.FieldByName(key.String())
				value, err := eval(elem.Value)
				if err != nil {
					return reflect.Value{}, err
				}
				field.Set(value.Convert(field.Type()))
			}
		}
		return rv, nil
	default:
		return t, nil
	}
}

func evalSelectorExpr(expr *ast.SelectorExpr) (reflect.Value, error) {
	x, err := eval(expr.X)
	if err != nil {
		return reflect.Value{}, err
	}
	sel, err := eval(expr.Sel)
	if err != nil {
		return reflect.Value{}, err
	}
	x = reflect.Indirect(x)
	switch x.Kind() {
	case reflect.String:
		if pkg, ok := packages()[x.String()]; ok {
			if fn, ok := pkg[sel.String()]; ok {
				return fn, nil
			}
			return reflect.Value{}, fmt.Errorf("cannot find object in package: %s.%s", x, sel)
		}
		return reflect.Value{}, fmt.Errorf("cannot find package: %s", x)
	case reflect.Struct:
		method := x.MethodByName(sel.String())
		if method.IsValid() {
			return method, nil
		}
		field := x.FieldByName(sel.String())
		if field.IsValid() {
			return field, nil
		}
		return reflect.Value{}, fmt.Errorf("no field or method %q on %v", sel.String(), x.Type())
	}
	return reflect.Value{}, fmt.Errorf("unsupported *ast.SelectorExpr: (%v)(%v)", x.Type(), x)
}

func evalUnaryExpr(expr *ast.UnaryExpr) (reflect.Value, error) {
	v, err := eval(expr.X)
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
		return reflect.ValueOf(!v.Bool()), nil
	case token.XOR:
		return reflect.ValueOf(^v.Int()), nil
	}
	return reflect.Value{}, fmt.Errorf("unsupported *ast.UnaryExpr: %T", expr.Op)
}

func evalIndexExpr(expr *ast.IndexExpr) (reflect.Value, error) {
	return reflect.Value{}, fmt.Errorf("*ast.IndexExpr unsupported")
}

func evalSliceExpr(expr *ast.SliceExpr) (reflect.Value, error) {
	return reflect.Value{}, fmt.Errorf("*ast.SliceExpr unsupported")
}
