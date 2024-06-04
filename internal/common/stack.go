package common

import (
	"fmt"
	"strings"
)

func MakeUid(name string, nbt *string) string {
	if nbt == nil {
		return name
	}
	return fmt.Sprintf("%s:%s", name, *nbt)
}

func FromUid(uid string) (string, *string) {
	parts := strings.Split(uid, ":")
	if len(parts) == 3 && len(parts[2]) == 32 {
		return parts[0] + ":" + parts[1], &parts[2]
	}
	return uid, nil
}
