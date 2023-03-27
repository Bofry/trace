package trace

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
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

func OS() KeyValue {
	if len(__attrval_os) == 0 {
		__attrval_os = runtime.GOOS + "-" + runtime.GOARCH
	}
	return __ATTR_OS.String(__attrval_os)
}

func Pid() KeyValue {
	if __attrval_pid == 0 {
		__attrval_pid = os.Getpid()
	}
	return __ATTR_PID.Int(__attrval_pid)
}

func Facility(v string) KeyValue {
	return __ATTR_FACILITY.String(v)
}

func Signature(v string) KeyValue {
	return __ATTR_SIGNATURE.String(v)
}

func Version(v string) KeyValue {
	return __ATTR_VERSION.String(v)
}

func HttpMethod(v string) KeyValue {
	return __ATTR_HTTP_METHOD.String(v)
}

func HttpRequest(v string) KeyValue {
	return __ATTR_HTTP_REQUEST.String(v)
}

func HttpRequestPath(v string) KeyValue {
	return __ATTR_HTTP_REQUEST_PATH.String(v)
}

func HttpResponse(v string) KeyValue {
	return __ATTR_HTTP_RESPONSE.String(v)
}

func HttpUserAgent(v string) KeyValue {
	return __ATTR_HTTP_USER_AGENT.String(v)
}

func Stringer(name string, o interface{}) KeyValue {
	return stringer(Key(name), o)
}

func Infer(name string, s interface{}) KeyValue {
	return infer(Key(name), s)
}

func expandObject(namespace string, o interface{}) []KeyValue {
	// value is struct, bool, string, int, int64, float64?
	switch v := o.(type) {
	case TracerTagMarshaler:
		builder := &TracerTagBuilder{
			namespace: namespace,
		}
		err := v.MarshalTracerTag(builder)
		if err != nil {
			return []KeyValue{
				Key(fmt.Sprintf("%s_error", namespace)).String(err.Error()),
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
				attrkey := Key(namespace + "." + k)

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

func stringer(key Key, o interface{}) KeyValue {
	switch v := o.(type) {
	case fmt.Stringer:
		return key.String(v.String())
	}
	return key.String(fmt.Sprintf("%+v", o))
}

func infer(key Key, s interface{}) KeyValue {
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
