package example

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

func ExampleFunction() {
	fmt.Println("Example Go program")

	// Use some imported packages
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	if len(os.Args) > 1 {
		fmt.Printf("Args: %v\n", os.Args[1:])
	}

	// Reference websocket to show external dependency
	upgrader := websocket.Upgrader{}
	fmt.Printf("Upgrader: %+v\n", upgrader)
}
