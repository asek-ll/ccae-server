package common

import (
	"fmt"
	"strings"
)

var customNbt = make(map[string]string)
var nbtReplace = make(map[string]string)

func init() {
	customNbt["pneumaticcraft:liquid_compressor"] = ""
	customNbt["pneumaticcraft:advanced_liquid_compressor"] = ""
	customNbt["bloodmagic:soulgempetty"] = "e523ac9950fc5deae9f856f304481b2e"
	customNbt["bloodmagic:soulgemlesser"] = "6b919a4b9544aeea009b20dc5162e16a"
	customNbt["occultism:satchel"] = "ef92b33b731066e4c3d98dbbcbc40b80"
	customNbt["enderstorage:ender_chest"] = ""
	customNbt["astralsorcery:rock_crystal"] = "b9c70089b3f99aafa3164a74864ad8ca"
	customNbt["astralsorcery:celestial_gateway"] = "b9c70089b3f99aafa3164a74864ad8ca"

	nbtReplace["fa498b90c4fe78d5a4e7185cfbeecb99"] = "b729bba220cd3bbe2881bb0f71b31f54"
	nbtReplace["21326e7bd59842698f7ea18b0b3d8a7e"] = "c79f7e2cdc4552303fb8e490b6e3f958"
}

func MakeUid(name string, nbt *string) string {
	if correctNbt, e := customNbt[name]; e {
		if correctNbt != "" {
			return fmt.Sprintf("%s:%s", name, correctNbt)
		}
		return name
	}

	if nbt == nil {
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
