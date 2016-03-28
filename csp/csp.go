package csp

import (
	"reflect"
)

type Process interface {
	Run()
}

// Pass addresses to two channels.
func Connect(a, b interface{}) {
	ConnectBuf(a, b, 0)
}

func ConnectBuf(a, b interface{}, size int) {
	aVal := reflect.ValueOf(a).Elem()
	bVal := reflect.ValueOf(b).Elem()

	aType := aVal.Type()
	bType := bVal.Type()

	// Check that at least one endpoint is not connected.
	if !aVal.IsNil() && !bVal.IsNil() {
		panic("Endpoints already connected.")
	}

	// Check that the endpoints have the same type.
	if aType != bType {
		panic("Endpoint types don't match.")
	}

	if !aVal.IsNil() {
		bVal.Set(aVal)
		return
	} else if !bVal.IsNil() {
		aVal.Set(bVal)
		return
	}

	aVal.Set(reflect.MakeChan(aType, size))
	bVal.Set(aVal)
}

// Call `Run` on each Process. The last process is run in the foreground.
func Run(procs ...Process) {
	for i := 0; i < len(procs)-1; i++ {
		go procs[i].Run()
	}
	procs[len(procs)-1].Run()
}
