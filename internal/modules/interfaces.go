package modules

// KeyReader interface for reading user input
type KeyReader interface {
	ReadKey() (string, error)
}

// Writer interface for output operations
type Writer interface {
	Write([]byte) (int, error)
}
