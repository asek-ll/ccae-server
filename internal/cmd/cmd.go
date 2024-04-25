package cmd

type Command interface {
	Exec(args []string) error
}
