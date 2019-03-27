package prompt

type Prompter interface {
	Message(string) error

	Prompt(string) (string, error)
}
