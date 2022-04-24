package main

import (
	"fmt"
	"time"

	"github.com/Ishan27g/go-utils/noop/example"
)

func main() {

	s1, s2, s3, s4 := example.Setup()
	example.SetNoop(false)

	fmt.Println("TriggerAsync---")
	example.TriggerAsync(s1, s2, s3, s4) // normal async-execution
	<-time.After(6 * time.Second)
	fmt.Println("\nSendHttpNoop noop=false ---")
	example.SendHttpNoop(s1, s2, s3, s4, false) // triggering with `noop=false` should have the same effect as async

	<-time.After(6 * time.Second)
	example.SetNoop(true)
	fmt.Println("\nSendHttpNoop noop=true ---")
	example.SendHttpNoop(s1, s2, s3, s4, true) // triggering with `noop=true`

}
