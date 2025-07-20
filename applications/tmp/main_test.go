package main

import (
	"context"
	"testing"
)

func TestHandler(t *testing.T) {
	ctx := context.Background()

	result, err := handler(ctx)
	if err != nil {
		t.Errorf("handler() returned an error: %v", err)
	}

	if result != "tmp2" {
		t.Errorf("handler() returned %q, expected %q", result, "tmp")
	}
}
