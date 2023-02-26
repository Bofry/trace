package trace

import "fmt"

var (
	_ TracerTagMarshaler = VarsBuilder(make(map[string]interface{}))
)

type VarsBuilder map[string]interface{}

func (b VarsBuilder) Put(key string, value interface{}) VarsBuilder {
	if _, ok := b[key]; ok {
		panic(
			fmt.Sprintf("Cannot add duplicate key '%s' into %T", key, b))
	}
	b[key] = value
	return b
}

func (b VarsBuilder) Set(key string, value interface{}) VarsBuilder {
	b[key] = value
	return b
}

func (b VarsBuilder) Delete(key string) VarsBuilder {
	delete(b, key)
	return b
}

func (b VarsBuilder) MarshalTracerTag(builder *TracerTagBuilder) error {
	for k, v := range b {
		builder.Value(k, v)
	}
	return nil
}
