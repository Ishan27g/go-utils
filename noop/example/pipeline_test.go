package example

import (
	"fmt"
	"testing"
)

func Test_Setup(t *testing.T) {
	fmt.Println("noop = false")

	isNoop = false
	RunAsync()

	fmt.Println("\nnoop = true")

	isNoop = true
	RunSync()
}
