package common

func MapKeys[K comparable, V any](m map[K]V) []K {
	var values []K
	for k := range m {
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

func CopyMap[K comparable, V any](m map[K]V) map[K]V {
	cp := make(map[K]V)
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func Unique[K comparable, V any](s []V, key func(v V) K) []V {
	set := make(map[K]struct{})
	j := 0

	for _, i := range s {
		k := key(i)
		if _, e := set[k]; !e {
			s[j] = i
			set[k] = struct{}{}
			j += 1
		}
	}
	return s[0:j]
}
