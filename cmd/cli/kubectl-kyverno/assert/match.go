package assert

import (
	"context"
	"fmt"
	reflectutils "github.com/kyverno/kyverno-json/pkg/utils/reflect"
	"reflect"
)

func GetKind(value any) reflect.Kind {
	if value == nil {
		return reflect.Invalid
	}
	return reflect.TypeOf(value).Kind()
}

func Match(ctx context.Context, expected, actual any) (bool, error) {
	if expected != nil {
		switch reflectutils.GetKind(expected) {
		case reflect.Slice:
			if reflectutils.GetKind(actual) != reflect.Slice {
				return false, fmt.Errorf("invalid actual value, must be a slice, found %s", reflectutils.GetKind(actual))
			}
			if reflect.ValueOf(expected).Len() != reflect.ValueOf(actual).Len() {
				return false, nil
			}
			for i := 0; i < reflect.ValueOf(expected).Len(); i++ {
				if inner, err := Match(ctx, reflect.ValueOf(expected).Index(i).Interface(), reflect.ValueOf(actual).Index(i).Interface()); err != nil {
					return false, err
				} else if !inner {
					return false, nil
				}
			}
			return true, nil
		case reflect.Map:
			if reflectutils.GetKind(actual) != reflect.Map {
				return false, fmt.Errorf("invalid actual value, must be a map, found %s", reflectutils.GetKind(actual))
			}
			iter := reflect.ValueOf(expected).MapRange()
			for iter.Next() {
				actualValue := reflect.ValueOf(actual).MapIndex(iter.Key())
				if !actualValue.IsValid() {
					return false, nil
				}
				if inner, err := Match(ctx, iter.Value().Interface(), actualValue.Interface()); err != nil {
					return false, err
				} else if !inner {
					return false, nil
				}
			}
			return true, nil
		}
	}
	return reflectutils.MatchScalar(expected, actual)
}

func MatchScalar(expected, actual any) (bool, error) {
	if actual == nil && expected == nil {
		return true, nil
	} else if actual == nil && expected != nil {
		return false, nil
	} else if actual != nil && expected == nil {
		return false, nil
	}
	if actual == expected {
		return true, nil
	}
	// if they are the same type we can use reflect.DeepEqual
	if reflect.TypeOf(expected) == reflect.TypeOf(actual) {
		return reflect.DeepEqual(expected, actual), nil
	}
	e := reflect.ValueOf(expected)
	a := reflect.ValueOf(actual)
	if !a.IsValid() && !e.IsValid() {
		return true, nil
	}
	if a.CanComplex() && e.CanComplex() {
		return a.Complex() == e.Complex(), nil
	}
	if a.CanFloat() && e.CanFloat() {
		return a.Float() == e.Float(), nil
	}
	if a.CanInt() && e.CanInt() {
		return a.Int() == e.Int(), nil
	}
	if a.CanUint() && e.CanUint() {
		return a.Uint() == e.Uint(), nil
	}
	if a, ok := ToNumber(a); ok {
		if e, ok := ToNumber(e); ok {
			return a == e, nil
		}
	}
	return false, fmt.Errorf("types are not comparable, %s - %s", GetKind(expected), GetKind(actual))
}

func ToNumber(value reflect.Value) (float64, bool) {
	if value.CanFloat() {
		return value.Float(), true
	}
	if value.CanInt() {
		return float64(value.Int()), true
	}
	if value.CanUint() {
		return float64(value.Uint()), true
	}
	return 0, false
}
