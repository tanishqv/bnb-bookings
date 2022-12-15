package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	var mh myHandler

	h := NoSurf(&mh)
	switch v := h.(type) {
	case http.Handler:
		// Nothing
	default:
		t.Errorf(fmt.Sprintf("type is not http.Handler, but is %T", v))
	}
}

func TestSessionLoad(t *testing.T) {
	var mh myHandler

	h := SessionLoad(&mh)
	switch v := h.(type) {
	case http.Handler:
		// Do nothing; test passed
	default:
		t.Errorf(fmt.Sprintf("type is not http.Handler, but is %T", v))
	}
}
