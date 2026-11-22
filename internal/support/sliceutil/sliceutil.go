// Package sliceutil provides small generic slice helpers.
package sliceutil

// AppendUnique appends values that are not already present.
func AppendUnique[T comparable](values []T, add ...T) []T {
	seen := make(map[T]struct{}, len(values)+len(add))
	for _, value := range values {
		seen[value] = struct{}{}
	}
	for _, value := range add {
		if _, ok := seen[value]; ok {
			continue
		}
		values = append(values, value)
		seen[value] = struct{}{}
	}
	return values
}
