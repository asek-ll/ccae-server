package common

import (
	"fmt"
	"strings"
)

func MakeUid(name string, nbt *string) string {
	if nbt == nil {
		return name
	}
	if *nbt == "fa498b90c4fe78d5a4e7185cfbeecb99" {
		return fmt.Sprintf("%s:b729bba220cd3bbe2881bb0f71b31f54", name)
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

func IsFluid(uid string) bool {
	return strings.HasPrefix(uid, "fluid:")
}

func FluidUid(uid string) string {
	return strings.TrimPrefix(uid, "fluid:")
}
