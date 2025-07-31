package modules

import (
	"golang.org/x/term"
)

// Module interface defines the contract for all BBS modules
type Module interface {
	// Execute runs the module and returns true if the session should continue
	Execute(term *term.Terminal) bool
}
