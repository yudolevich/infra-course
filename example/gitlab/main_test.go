package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	http.HandlerFunc(Handler).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf(
			"wrong status code: want %v got %v",
			http.StatusOK, rec.Code,
		)
	}

	if rec.Body.String() != "Hello!\n" {
		t.Errorf("wrong body: got %s", rec.Body.String())
	}
}
