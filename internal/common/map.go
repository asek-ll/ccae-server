package common

func MapKeys[K comparable, V any](m map[K]V) []K {
	var values []K
	for k, _ := range m {
		values = append(values, k)
	}
	return values
}

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

func ToSet[K comparable](s []K) map[K]struct{} {
	set := make(map[K]struct{})

	for _, i := range s {
		set[i] = struct{}{}
	}

	return set
}
