package common

import (
	"encoding/base64"
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

	customNbt["bloodmagic:firescribetool"] = "552887824c43124013fd24f6edcde0fb"
	customNbt["bloodmagic:airscribetool"] = "552887824c43124013fd24f6edcde0fb"
	customNbt["bloodmagic:earthscribetool"] = "552887824c43124013fd24f6edcde0fb"
	customNbt["bloodmagic:waterscribetool"] = "552887824c43124013fd24f6edcde0fb"

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

var QuestMarkIcon, _ = base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAMAAABg3Am1AAAB+FBMVEUvLy8vLy8wMDAvLy8wMDAsLCwwMDBHcEwwMDAvLy8vLy8wMDAvLy8vLy8wMDAvLy8sLCwvLy8vLy8vLy8vLy8uLi4tLS0vLy8vLy8vLy8vLy8wMDAsLCwwMDAvLy8wMDAvLy8wMDAwMDAnJycvLy8wMDAvLy8wMDAvLy8vLy8wMDAwMDAvLy8vLy8vLy8wMDAwMDAvLy8wMDAvLy8vLy8rKysvLy8vLy8vLy8vLy8wMDAwMDAwMDAwMDAvLy8qKiovLy8vLy8wMDAwMDAvLy8rKyswMDAtLS0wMDAwMDAwMDAwMDAvLy8uLi4wMDAvLy8vLy8wMDAwMDAwMDAvLy8vLy8wMDAwMDAvLy8wMDAwMDAvLy8uLi4wMDAvLy8tLS0vLy8wMDAwMDAvLy8wMDAvLy8wMDAvLy8vLy8sLCwwMDAvLy8vLy8vLy8wMDAvLy8wMDAvLy8wMDAuLi4vLy8qKiovLy8wMDAwMDAwMDAvLy8vLy8wMDAvLy8qKiowMDAvLy8rKyswMDAvLy8vLy8vLy8vLy8vLy8vLy8wMDAwMDAwMDAvLy8qKiowMDAtLS0vLy8vLy8wMDAvLy8vLy8wMDAvLy8vLy8vLy8vLy8vLy8vLy8vLy8wMDAvLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8wMDB8kbe9AAAAp3RSTlP+9/ss/QL4AP79+lCxnG4aA7T8Zs0IBF17QijsAdKos/BWGAL1W7fCNtuEZZ1+HGrwIZPLrCPvT/QwO36iQLsI333qiF43ywpK7yvQ6RpPhyXV+dpO1DrAlqHF3g9ooAmBjCpibUOwhskI1Clsr7rym9PeCXYGOMmq7R7n8sgVtIoBdq4tDpTXdeKAHvsFhQcdwlHM1mCiK5K2G+4qNXyLznGp0hM85ideAd8AAAH+SURBVEjHY2DHAHMWLlo8m5Ohx8c4TrgKQ5YBjS82dQHHcoblMMQSzi2DV4PkFAZ0UGuJW4NYIudyJOPBaDkDmz4uDaKC6KrBGpZzOGLXwMrDgAuIY9VgDTQM5tdgu/ooBZgNyxlqurFoSAqCaWBL1wMJ9HJnwTQst8GioRlmP2MqTCg+FybGxY+hwZBlOdQGPoQrM2Vhfs/B0FCyHKpBhBUpTPxgGuQwNMCDyBw51O0ZoaIiGBoKoTYUsaJEbCfUBk0MDe1QDVoo6pU9oRpCMTTwQqzmzEbRUAZzqBqGhglOYBsKUBNjHszTCZjxwOzGwcCkI4qivgMWcRwt2BKfqWsdqvmxQjAN8jgzEDKQgCc+iwAiNAhEcMATnzY7YQ3FafD8sLyBnbCG1omI3ODPTlhDTAY8x3E2sRPW0K8Oz6JMyexEaGiDO0clhJ0IDRossLwqHcZOjAZBmHuqvdiJ0eArBHWPGTM7URpUoe7hUGQnTkMfVIM3O5EaoFmDQZdYDZFQG2yJ1eAM0cDlTqyGpRAXzWQnVgP7fFDClp5OvAZ2yRnT5s1iJ0EDHkAlDUsmic+VIV5DowSoRJ0sRbSGLkg8LFMiUgM/GzSmjYjU4ALLbdFEauCDZbdKIjV4wDSUE6mh1AHqJGFiQ8kqH2yDCSvRESdVYcAZmCKANR4AylXnkqHv7kAAAAAASUVORK5CYII=")
