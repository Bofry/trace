package trace

type TracerTagBuilder struct {
	namespace string
	container []KeyValue
}

func (builder *TracerTagBuilder) Value(name string, v any) {
	if len(name) > 0 {
		k := builder.namespaceKey(name)
		kv := expandObject(string(k), v)
		if len(kv) > 0 {
			builder.container = append(builder.container, kv...)
		}
	}
}

func (builder *TracerTagBuilder) String(name string, v string) {
	if len(name) > 0 {
		kv := builder.namespaceKey(name).String(v)
		builder.container = append(builder.container, kv)
	}
}

func (builder *TracerTagBuilder) StringSlice(name string, v []string) {
	if len(name) > 0 {
		kv := builder.namespaceKey(name).StringSlice(v)
		builder.container = append(builder.container, kv)
	}
}

func (builder *TracerTagBuilder) Float64(name string, v float64) {
	if len(name) > 0 {
		kv := builder.namespaceKey(name).Float64(v)
		builder.container = append(builder.container, kv)
	}
}

func (builder *TracerTagBuilder) Float64Slice(name string, v []float64) {
	if len(name) > 0 {
		kv := builder.namespaceKey(name).Float64Slice(v)
		builder.container = append(builder.container, kv)
	}
}

func (builder *TracerTagBuilder) Int(name string, v int) {
	if len(name) > 0 {
		kv := builder.namespaceKey(name).Int(v)
		builder.container = append(builder.container, kv)
	}
}

func (builder *TracerTagBuilder) IntSlice(name string, v []int) {
	if len(name) > 0 {
		kv := builder.namespaceKey(name).IntSlice(v)
		builder.container = append(builder.container, kv)
	}
}

func (builder *TracerTagBuilder) Int64(name string, v int64) {
	if len(name) > 0 {
		kv := builder.namespaceKey(name).Int64(v)
		builder.container = append(builder.container, kv)
	}
}

func (builder *TracerTagBuilder) Int64Slice(name string, v []int64) {
	if len(name) > 0 {
		kv := builder.namespaceKey(name).Int64Slice(v)
		builder.container = append(builder.container, kv)
	}
}

func (builder *TracerTagBuilder) Bool(name string, v bool) {
	if len(name) > 0 {
		kv := builder.namespaceKey(name).Bool(v)
		builder.container = append(builder.container, kv)
	}
}

func (builder *TracerTagBuilder) BoolSlice(name string, v []bool) {
	if len(name) > 0 {
		kv := builder.namespaceKey(name).BoolSlice(v)
		builder.container = append(builder.container, kv)
	}
}

func (builder *TracerTagBuilder) Result() []KeyValue {
	return builder.container
}

func (builder *TracerTagBuilder) namespaceKey(name string) Key {
	return Key(builder.namespace + "." + name)
}
