package main

import (
        "log"
        "net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello!\n"))
}

func main() {
        if err := http.ListenAndServe("0.0.0.0:8080", http.HandlerFunc(Handler));err != nil {
                log.Fatal(err)
        }
}
