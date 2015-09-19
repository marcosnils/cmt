package cmd

type Cmd interface {
	Run(c string, args ...string) (string, string, error)
}
