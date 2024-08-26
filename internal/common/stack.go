package common

import (
	"fmt"
	"strings"
)

var anynbt = make(map[string]struct{})
var nbtReplace = make(map[string]string)

func init() {
	anynbt["pneumaticcraft:liquid_compressor"] = struct{}{}
	anynbt["pneumaticcraft:advanced_liquid_compressor"] = struct{}{}

	nbtReplace["fa498b90c4fe78d5a4e7185cfbeecb99"] = "b729bba220cd3bbe2881bb0f71b31f54"
	nbtReplace["21326e7bd59842698f7ea18b0b3d8a7e"] = "c79f7e2cdc4552303fb8e490b6e3f958"
}

func MakeUid(name string, nbt *string) string {
	if nbt == nil {
		return name
	}

	if _, e := anynbt[name]; e {
		return name
	}

	if correctNbt, e := nbtReplace[*nbt]; e {
		return fmt.Sprintf("%s:%s", name, correctNbt)
	}

	// if *nbt == "fa498b90c4fe78d5a4e7185cfbeecb99" {
	// 	return fmt.Sprintf("%s:b729bba220cd3bbe2881bb0f71b31f54", name)
	// }

	// if *nbt == "21326e7bd59842698f7ea18b0b3d8a7e" {
	// 	return fmt.Sprintf("%s:c79f7e2cdc4552303fb8e490b6e3f958", name)
	// }
	// if name == "pneumaticcraft:liquid_compressor" {
	// 	return name
	// }

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
