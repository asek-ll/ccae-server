package crafter

type Step struct {
	Recipe  Recipe
	Repeats int
}

type Plan struct {
	Steps []Step
	Goals []Stack
}
