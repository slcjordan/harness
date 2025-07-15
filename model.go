package harness

type Stream uint8

const (
	Stdin Stream = iota
	Stdout
	Stderr
)

type CommandEvent struct {
	Stream Stream
	Data   []byte
}

type InteractiveInput struct {
	Stream Stream
	Data   []byte
}
