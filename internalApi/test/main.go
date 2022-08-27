package main

import (
	"github.com/Ishan27g/go-utils/tracing"
	"github.com/Ishan27g/internalApi/test/server"
)

func main() {
	provider := tracing.Init("jaeger", "serviceX", "users")
	defer func() {
		defer provider.Close()
	}()
	server.Run(provider)
}
