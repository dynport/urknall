package urknall

type Host interface {
	Command(cmd string) (Command, error)
	User() string
	String() string
}
