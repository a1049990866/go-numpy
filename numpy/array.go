package numpy

import (
	"errors"
	"math"
	"reflect"
	"regexp"
)

type Array struct {
	Value       interface{}
	Shape       []int
	Err         error
	dimensional int
}

func NewArray(value interface{}, args ...int) *Array {
	if reflect.TypeOf(value) == reflect.TypeOf(&Array{}) {
		return value.(*Array)
	}
	data := &Array{Value: value}
	data.Err = data.shape(reflect.ValueOf(value), args...)
	return data
}

func (a *Array) isNumber(t reflect.Kind) bool {
	s := t.String()
	reg, _ := regexp.Compile("(u?int\\d{0,2})|(float\\d\\d)")
	return reg.FindString(s) != ""
}

func (a *Array) interfaceToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		a.Err = errors.New("请传入数值类型")
		return 0
	}
}

func (a *Array) numberOperation(v1, v2 interface{}, op string) float64 {
	v3 := a.interfaceToFloat64(v1)
	v4 := a.interfaceToFloat64(v2)
	switch op {
	case "+":
		fallthrough
	case "add":
		fallthrough
	case "sum":
		return v3 + v4
	case "-":
		return v3 - v4
	case "*":
		return v3 * v4
	case "/":
		return v3 / v4
	case "//":
		return float64(int(v3 / v4))
	case "%":
		return float64(int(v3) % int(v4))
	case "pow":
		return math.Pow(v3, v4)
	default:
		a.Err = errors.New("指定的运算符不存在")
		return 0
	}
}

func (a *Array) shape(v reflect.Value, args ...int) error {
	if len(args) < 2 {
		args = append(args, 0, 0)
	}
	if a.dimensional == 0 {
		args[1] = 0
	}
	t := v.Type().Kind()
	if t != reflect.Slice {
		if a.dimensional == 0 {
			return errors.New("请传入数组")
		}
		if !a.isNumber(t) {
			return errors.New("数组元素只能是数值类型")
		}
		return nil
	}
	n := v.Len()
	if args[1] >= len(a.Shape) {
		a.Shape = append(a.Shape, n)
		a.dimensional += 1
	} else if a.Shape[args[1]] != n {
		return errors.New("数据长度不一致, 请检测传入的数据结构")
	}
	args[1] += 1
	for i := 0; i < n; i++ {
		if i != 0 && args[0] == 0 {
			return nil
		}
		err := a.shape(reflect.ValueOf(v.Index(i).Interface()), args[0], args[1])
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Array) calculate(v reflect.Value, n int, op string, res float64) float64 {
	if n >= a.dimensional {
		res = a.numberOperation(res, v.Interface(), op)
		return res
	}
	for i := 0; i < a.Shape[n]; i++ {
		res += a.calculate(reflect.ValueOf(v.Index(i).Interface()), n+1, op, 0)
	}
	return res
}

func (a *Array) Sum() float64 {
	return a.calculate(reflect.ValueOf(a.Value), 0, "+", 0)
}

func (a *Array) zero(shape []int, def interface{}) (res []interface{}) {
	n := len(shape)
	res = make([]interface{}, shape[0])
	for i := 0; i < shape[0]; i++ {
		if n == 1 {
			res[i] = def
		} else {
			res[i] = a.zero(shape[1:], def)
		}
	}
	return
}

func (a *Array) Zero(shape []int, def interface{}) *Array {
	return NewArray(a.zero(shape, def))
}

func (a *Array) set(data interface{}, v interface{}, position ...int) {
	vo := reflect.ValueOf(data)
	for i, p := range position {
		if i < a.dimensional-1 {
			vo = reflect.ValueOf(vo.Index(p).Interface())
		} else {
			vo.Index(p).Set(reflect.ValueOf(a.interfaceToFloat64(v)))
		}
	}
}

func (a *Array) t(data interface{}, res interface{}, axis, dim int, position ...int) {
	if dim == a.dimensional {
		pos := append([]int{position[axis]}, position[:axis]...)
		pos = append(pos, position[axis+1:]...)
		a.set(res, data, pos...)
		return
	}
	d := reflect.ValueOf(data)
	for i := 0; i < a.Shape[dim]; i++ {
		a.t(d.Index(i).Interface(), res, axis, dim+1, append(position, i)...)
	}
}

func (a *Array) T(axis int) *Array {
	shape := append([]int{a.Shape[axis]}, a.Shape[:axis]...)
	shape = append(shape, a.Shape[axis+1:]...)
	res := a.zero(shape, float64(0))
	a.t(a.Value, res, axis, 0)
	return NewArray(res)
}

func (a *Array) opArray(op string, data ...reflect.Value) {
	if data[0].Type().Kind() != reflect.Slice {
		for _, v := range data[1:] {
			num := a.numberOperation(data[0].Interface(), v.Interface(), op)
			data[0].Set(reflect.ValueOf(num))
		}
		return
	}
	for i := 0; i < data[0].Len(); i++ {
		var d2 []reflect.Value
		for _, v := range data {
			v1 := reflect.ValueOf(v.Index(i).Interface())
			if v1.Type().Kind() != reflect.Slice {
				v1 = v.Index(i)
			}
			d2 = append(d2, v1)
		}
		a.opArray(op, d2...)
	}
}

func (a *Array) getValues(data ...interface{}) []reflect.Value {
	d := []reflect.Value{reflect.ValueOf(a.Value)}
	for _, v := range data {
		if reflect.TypeOf(a) == reflect.TypeOf(v) {
			d = append(d, reflect.ValueOf(v.(*Array).Value))
		} else {
			d = append(d, reflect.ValueOf(v))
		}
	}
	return d
}

func (a *Array) Add(data ...interface{}) *Array {
	data = append([]interface{}{a.Value}, data...)
	res := a.Zero(a.Shape, 0)
	res.opArray("+", res.getValues(data...)...)
	return res
}

func (a *Array) Mul(data ...interface{}) *Array {
	data = append([]interface{}{a.Value}, data...)
	res := a.Zero(a.Shape, 1)
	res.opArray("*", res.getValues(data...)...)
	return res
}

func (a *Array) SumAxis(axis int) *Array {
	if a.dimensional > 1 {
		d := reflect.ValueOf(a.T(axis).Value)
		var val []reflect.Value
		for i := 0; i < d.Len(); i++ {
			val = append(val, reflect.ValueOf(d.Index(i).Interface()))
		}
		if len(val) == 0 {
			return a
		}
		a.opArray("+", val...)
		return NewArray(val[0].Interface())
	} else if a.dimensional == 1 {
		d := reflect.ValueOf(a.Value)
		var val []reflect.Value
		for i := 0; i < d.Len(); i++ {
			val = append(val, reflect.ValueOf(d.Index(i)))
		}
		if len(val) == 0 {
			return a
		}
		a.opArray("+", val...)
		return NewArray(val[0].Interface())
	} else {
		return a
	}
}

func (a *Array) opEveryElem(res, data interface{}, v float64, op string, dim int, position ...int) {
	if dim == a.dimensional {
		a.set(res, a.numberOperation(data, v, op), position...)
		return
	}
	d := reflect.ValueOf(data)
	for i := 0; i < a.Shape[dim]; i++ {
		a.opEveryElem(res, d.Index(i).Interface(), v, op, dim+1, append(position, i)...)
	}
}

func (a *Array) AddNumber(n interface{}) *Array {
	res := a.zero(a.Shape, float64(0))
	a.opEveryElem(res, a.Value, a.interfaceToFloat64(n), "+", 0)
	return NewArray(res)
}

func (a *Array) SubNumber(n interface{}) *Array {
	res := a.zero(a.Shape, float64(0))
	a.opEveryElem(res, a.Value, a.interfaceToFloat64(n), "-", 0)
	return NewArray(res)
}

func (a *Array) MulNumber(n interface{}) *Array {
	res := a.zero(a.Shape, float64(0))
	a.opEveryElem(res, a.Value, a.interfaceToFloat64(n), "*", 0)
	return NewArray(res)
}

func (a *Array) DivNumber(n interface{}) *Array {
	res := a.zero(a.Shape, float64(0))
	a.opEveryElem(res, a.Value, a.interfaceToFloat64(n), "/", 0)
	return NewArray(res)
}

// 整除
func (a *Array) DivisibleNumber(n interface{}) *Array {
	res := a.zero(a.Shape, float64(0))
	a.opEveryElem(res, a.Value, a.interfaceToFloat64(n), "//", 0)
	return NewArray(res)
}

//取余
func (a *Array) RemainderNumber(n interface{}) *Array {
	res := a.zero(a.Shape, float64(0))
	a.opEveryElem(res, a.Value, a.interfaceToFloat64(n), "%", 0)
	return NewArray(res)
}

func (a *Array) PowNumber(n interface{}) *Array {
	res := a.zero(a.Shape, float64(0))
	a.opEveryElem(res, a.Value, a.interfaceToFloat64(n), "pow", 0)
	return NewArray(res)
}
