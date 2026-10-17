package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, "Hello Fantasy Waiver Assistant!"); err != nil {
			fmt.Println("Error writing response:", err)
		}
	})

	fmt.Println("Server Running on 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
