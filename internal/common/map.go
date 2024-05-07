package common

func MapValues[K comparable, V any](m map[K]V) []V {
	var values []V
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func ToArgs[V any](s []V) []any {
	args := make([]any, len(s))
	for i, uid := range s {
		args[i] = uid
	}
	return args
}
