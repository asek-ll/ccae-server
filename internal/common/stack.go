package common

import "fmt"

func MakeUid(name string, nbt *string) string {
	if nbt == nil {
		return name
	}
	return fmt.Sprintf("%s:%s", name, *nbt)
}
