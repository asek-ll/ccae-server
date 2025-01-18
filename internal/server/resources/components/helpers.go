package components

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"

	"github.com/a-h/templ"
	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/server/handlers"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
)

var quest, _ = base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAMAAABg3Am1AAAB+FBMVEUvLy8vLy8wMDAvLy8wMDAsLCwwMDBHcEwwMDAvLy8vLy8wMDAvLy8vLy8wMDAvLy8sLCwvLy8vLy8vLy8vLy8uLi4tLS0vLy8vLy8vLy8vLy8wMDAsLCwwMDAvLy8wMDAvLy8wMDAwMDAnJycvLy8wMDAvLy8wMDAvLy8vLy8wMDAwMDAvLy8vLy8vLy8wMDAwMDAvLy8wMDAvLy8vLy8rKysvLy8vLy8vLy8vLy8wMDAwMDAwMDAwMDAvLy8qKiovLy8vLy8wMDAwMDAvLy8rKyswMDAtLS0wMDAwMDAwMDAwMDAvLy8uLi4wMDAvLy8vLy8wMDAwMDAwMDAvLy8vLy8wMDAwMDAvLy8wMDAwMDAvLy8uLi4wMDAvLy8tLS0vLy8wMDAwMDAvLy8wMDAvLy8wMDAvLy8vLy8sLCwwMDAvLy8vLy8vLy8wMDAvLy8wMDAvLy8wMDAuLi4vLy8qKiovLy8wMDAwMDAwMDAvLy8vLy8wMDAvLy8qKiowMDAvLy8rKyswMDAvLy8vLy8vLy8vLy8vLy8vLy8wMDAwMDAwMDAvLy8qKiowMDAtLS0vLy8vLy8wMDAvLy8vLy8wMDAvLy8vLy8vLy8vLy8vLy8vLy8vLy8wMDAvLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8wMDB8kbe9AAAAp3RSTlP+9/ss/QL4AP79+lCxnG4aA7T8Zs0IBF17QijsAdKos/BWGAL1W7fCNtuEZZ1+HGrwIZPLrCPvT/QwO36iQLsI333qiF43ywpK7yvQ6RpPhyXV+dpO1DrAlqHF3g9ooAmBjCpibUOwhskI1Clsr7rym9PeCXYGOMmq7R7n8sgVtIoBdq4tDpTXdeKAHvsFhQcdwlHM1mCiK5K2G+4qNXyLznGp0hM85ideAd8AAAH+SURBVEjHY2DHAHMWLlo8m5Ohx8c4TrgKQ5YBjS82dQHHcoblMMQSzi2DV4PkFAZ0UGuJW4NYIudyJOPBaDkDmz4uDaKC6KrBGpZzOGLXwMrDgAuIY9VgDTQM5tdgu/ooBZgNyxlqurFoSAqCaWBL1wMJ9HJnwTQst8GioRlmP2MqTCg+FybGxY+hwZBlOdQGPoQrM2Vhfs/B0FCyHKpBhBUpTPxgGuQwNMCDyBw51O0ZoaIiGBoKoTYUsaJEbCfUBk0MDe1QDVoo6pU9oRpCMTTwQqzmzEbRUAZzqBqGhglOYBsKUBNjHszTCZjxwOzGwcCkI4qivgMWcRwt2BKfqWsdqvmxQjAN8jgzEDKQgCc+iwAiNAhEcMATnzY7YQ3FafD8sLyBnbCG1omI3ODPTlhDTAY8x3E2sRPW0K8Oz6JMyexEaGiDO0clhJ0IDRossLwqHcZOjAZBmHuqvdiJ0eArBHWPGTM7URpUoe7hUGQnTkMfVIM3O5EaoFmDQZdYDZFQG2yJ1eAM0cDlTqyGpRAXzWQnVgP7fFDClp5OvAZ2yRnT5s1iJ0EDHkAlDUsmic+VIV5DowSoRJ0sRbSGLkg8LFMiUgM/GzSmjYjU4ALLbdFEauCDZbdKIjV4wDSUE6mh1AHqJGFiQ8kqH2yDCSvRESdVYcAZmCKANR4AylXnkqHv7kAAAAAASUVORK5CYII=")

func getItem(ctx context.Context, uid string) *dao.Item {
	items := ctx.Value("items").(map[string]*dao.Item)
	if items[uid] == nil {
		return &dao.Item{
			UID:         uid,
			DisplayName: uid,
			Icon:        quest,
		}
	}
	return items[uid]
}

func FormatIngredients(r *dao.Recipe) [][]*dao.RecipeItem {
	var rows [][]*dao.RecipeItem

	maxSlot := 0
	allWithSlot := true
	for _, ing := range r.Ingredients {
		if ing.Slot == nil {
			allWithSlot = false
			break
		}

		if *ing.Slot > maxSlot {
			maxSlot = *ing.Slot
		}
	}

	if allWithSlot {
		columnCount := int(math.Ceil(math.Sqrt(float64(maxSlot))))
		if columnCount < 3 {
			columnCount = 3
		}

		for i := 0; i < columnCount; i += 1 {
			rows = append(rows, make([]*dao.RecipeItem, columnCount))
		}

		for _, ri := range r.Ingredients {
			r := ((*ri.Slot) - 1) / columnCount
			c := ((*ri.Slot) - 1) % columnCount
			rows[r][c] = &ri
		}
	} else {
		columnCount := 3
		var currentRow []*dao.RecipeItem
		for i, ri := range r.Ingredients {
			if len(currentRow) == 0 {
				currentRow = make([]*dao.RecipeItem, columnCount)
				rows = append(rows, currentRow)
			}

			c := i % columnCount

			currentRow[c] = &ri

			if c == columnCount-1 {
				currentRow = nil
			}
		}
	}

	return rows
}

func GetRowCount(items int, width int) int {
	if items%width == 0 {
		return items / width
	}
	return items/width + 1
}

func ToTable[T any](items []T, width int) [][]T {
	var rows [][]T
	for i := 0; i < len(items); i += width {
		end := i + width
		if end > len(items) {
			end = len(items)
		}
		rows = append(rows, items[i:end])
	}
	return rows
}

func RecipeItemsToParams(items []dao.RecipeItem) url.Values {
	params := url.Values(make(map[string][]string))
	for slot, item := range items {
		if item.Slot != nil {
			params.Set(fmt.Sprintf("slot_%d", slot), fmt.Sprintf("%d", *item.Slot))
		}
		params.Set(fmt.Sprintf("item_%d", slot), item.ItemUID)
		params.Set(fmt.Sprintf("role_%d", slot), item.Role)
		params.Set(fmt.Sprintf("amount_%d", slot), strconv.Itoa(item.Amount))
	}

	return params
}

func RecipeToURL(recipe *dao.Recipe) string {
	params := url.Values(make(map[string][]string))
	params.Set("name", recipe.Name)

	for slot, item := range recipe.Ingredients {
		params.Set(fmt.Sprintf("slot_%d", 100+slot), fmt.Sprintf("%d", *item.Slot))
		params.Set(fmt.Sprintf("item_%d", 100+slot), item.ItemUID)
		params.Set(fmt.Sprintf("role_%d", 100+slot), item.Role)
		params.Set(fmt.Sprintf("amount_%d", 100+slot), strconv.Itoa(item.Amount))
	}

	for slot, item := range recipe.Results {
		params.Set(fmt.Sprintf("slot_%d", slot), fmt.Sprintf("%d", slot+1))
		params.Set(fmt.Sprintf("item_%d", slot), item.ItemUID)
		params.Set(fmt.Sprintf("role_%d", slot), item.Role)
		params.Set(fmt.Sprintf("amount_%d", slot), strconv.Itoa(item.Amount))
	}

	return fmt.Sprintf("/recipes/new/?%s", url.Values(params).Encode())
}

func toPlanUrl(recipe *dao.Recipe) string {
	params := RecipeItemsToParams(recipe.Ingredients)
	return fmt.Sprintf("/craft-plans/new/?%s", params.Encode())
}

func itemJsonByUid(ctx context.Context, itemUid string) string {
	if itemUid == "" {
		return ""
	}
	item := getItem(ctx, itemUid)
	dto := handlers.ItemToDto(item)
	result, err := json.Marshal(dto)
	if err != nil {
		return ""
	}

	return string(result)
}

func formatItemStackAmount(amount int) string {
	if amount < 10_000 {
		return fmt.Sprintf("%d", amount)
	}
	if amount < 1_000_000 {
		return fmt.Sprintf("%dK", amount/1_000)
	}
	if amount < 1_000_000_000 {
		return fmt.Sprintf("%dM", amount/1_000_000)
	}
	return "∞"
}

func formatFluidStackAmount(amount int) string {
	if amount < 100 {
		return fmt.Sprintf("%dmb", amount)
	}
	if amount < 1_000_000 {
		return fmt.Sprintf("%sB", limitedString(amount, 1_000, 3))
	}
	if amount < 1_000_000_000 {
		return fmt.Sprintf("%sKB", limitedString(amount, 1_000_000, 2))
	}
	return fmt.Sprintf("%d", amount)
}

func formatStackAmount(amount int, uid string) string {
	if common.IsFluid(uid) {
		return formatFluidStackAmount(amount)
	}
	return formatItemStackAmount(amount)
}

func limitedString(amount int, bound int, limit int) string {
	result := amount / bound
	part := strconv.Itoa(result)

	secondPartSize := limit - len(part) - 1
	if secondPartSize <= 0 {
		return part
	}

	mult := int(math.Pow10(secondPartSize))
	decimalPart := (amount - result*bound) / mult
	if decimalPart == 0 {
		return part
	}
	for decimalPart%10 == 0 {
		decimalPart /= 10
	}
	return part + "." + strconv.Itoa(decimalPart)
}

func mapToComponents[T any](items []T, f func(item T) templ.ComponentFunc) []templ.Component {
	result := make([]templ.Component, len(items))
	for i, item := range items {
		result[i] = f(item)
	}
	return result
}

func craftPlanConsumed(plan *crafter.Plan) []templ.Component {
	var result []templ.Component
	for _, related := range plan.Related {
		consumed := min(related.Consumed-related.Produced, related.StorageAmount)
		if consumed > 0 {
			result = append(result, ItemStack(related.UID, consumed))
		}
	}
	return result
}

func craftPlanMissing(plan *crafter.Plan) []templ.Component {
	var result []templ.Component
	for _, related := range plan.Related {
		missing := related.Consumed - related.Produced - related.StorageAmount
		if missing > 0 {
			result = append(result, ItemStack(related.UID, missing))
		}
	}
	return result
}

func craftPlanCreated(plan *crafter.Plan) []templ.Component {
	var result []templ.Component
	for _, related := range plan.Related {
		created := related.Produced - related.Consumed
		if created > 0 {
			result = append(result, ItemStack(related.UID, created))
		}
	}
	return result
}
