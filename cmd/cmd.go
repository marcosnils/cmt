package cmd

type Cmd interface {
	Run(c string) (string, string, error)
}
