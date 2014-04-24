package urknall

type Commander interface {
	Command(cmd string) (Command, error)
}
