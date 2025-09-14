package trace

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

var (
	__attrval_os  string
	__attrval_pid int
)

func ServiceName(v string) KeyValue {
	return __ATTR_SERVICE_NAME.String(v)
}

func Environment(v string) KeyValue {
	return __ATTR_ENVIRONMENT.String(v)
}


func Pid() KeyValue {
	if __attrval_pid == 0 {
		__attrval_pid = os.Getpid()
	}
	return __ATTR_PID.Int(__attrval_pid)
}


func expandObject(namespace string, o any) []KeyValue {
	// Fast path for common types to avoid reflection
	switch v := o.(type) {
	case string:
		return []KeyValue{Key(namespace).String(v)}
	case int:
		return []KeyValue{Key(namespace).Int(v)}
	case int64:
		return []KeyValue{Key(namespace).Int64(v)}
	case float64:
		return []KeyValue{Key(namespace).Float64(v)}
	case bool:
		return []KeyValue{Key(namespace).Bool(v)}
	case []string:
		return []KeyValue{Key(namespace).StringSlice(v)}
	case []int:
		return []KeyValue{Key(namespace).IntSlice(v)}
	case []int64:
		return []KeyValue{Key(namespace).Int64Slice(v)}
	case []float64:
		return []KeyValue{Key(namespace).Float64Slice(v)}
	case []bool:
		return []KeyValue{Key(namespace).BoolSlice(v)}
	case TracerTagMarshaler:
		builder := &TracerTagBuilder{
			namespace: namespace,
			container: make([]KeyValue, 0, 8),
		}
		err := v.MarshalTracerTag(builder)
		if err != nil {
			var sb strings.Builder
			sb.Grow(len(namespace) + 6) // "_error" = 6 chars
			sb.WriteString(namespace)
			sb.WriteString("_error")
			return []KeyValue{
				Key(sb.String()).String(err.Error()),
			}
		}
		return builder.Result()
	default:
		kv := infer(Key(namespace), o)
		if kv.Valid() {
			return []KeyValue{
				kv,
			}
		}
	}

	// value is map?
	rv := reflect.ValueOf(o)
	switch rv.Kind() {
	case reflect.Map:
		if rv.Type().Key().Kind() == reflect.String {
			var mapkv []KeyValue

			iter := rv.MapRange()
			for iter.Next() {
				k := iter.Key().String()
				v := iter.Value().Interface()
				var sb strings.Builder
				sb.Grow(len(namespace) + 1 + len(k))
				sb.WriteString(namespace)
				sb.WriteByte('.')
				sb.WriteString(k)
				attrkey := Key(sb.String())

				kv := infer(attrkey, v)
				if kv.Valid() {
					mapkv = append(mapkv, kv)
				} else {
					kv = stringer(attrkey, v)
					mapkv = append(mapkv, kv)
				}
			}
			return mapkv
		}
	}

	// otherwise
	{
		kv := stringer(Key(namespace), o)
		if kv.Valid() {
			return []KeyValue{
				kv,
			}
		}
	}
	return nil
}

func stringer(key Key, o any) KeyValue {
	switch v := o.(type) {
	case fmt.Stringer:
		return key.String(v.String())
	}
	return key.String(fmt.Sprintf("%+v", o))
}

func infer(key Key, s any) KeyValue {
	switch v := s.(type) {
	case string:
		return key.String(v)
	case float64:
		return key.Float64(v)
	case int64:
		return key.Int64(v)
	case int:
		return key.Int(v)
	case bool:
		return key.Bool(v)
	case []string:
		return key.StringSlice(v)
	case []float64:
		return key.Float64Slice(v)
	case []int64:
		return key.Int64Slice(v)
	case []int:
		return key.IntSlice(v)
	case []bool:
		return key.BoolSlice(v)
	}
	return __InvalidKeyValue
}
