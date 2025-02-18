package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("GET /api/{player}", appHandler(APIarchiveListGet))
	mux.Handle("POST /api/{player}", appHandler(APIarchiveListPost))
	mux.Handle("GET /api/{player}/{archive}", appHandler(APIarchiveDataGet))
	mux.Handle("POST /api/{player}/{archive}", appHandler(APIarchiveDataPost))
	mux.Handle("GET /api/{player}/{archive}/{uuid}", appHandler(APIresultGet))
	mux.Handle("POST /api/{player}/{archive}/{uuid}", appHandler(APIresultPost))

	fmt.Fprintln(os.Stderr, "API Listening :24377/tcp") // 24377 = 'chess' in T9
	err := http.ListenAndServe(":24377", mux)           // all interfaces
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
