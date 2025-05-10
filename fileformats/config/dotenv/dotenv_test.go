package dotenv

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	_, err := Load(".missing.env")
	if err == nil {
		t.Error("Expected error loading missing file")
		return
	}

	data, err := Load("testdata/.env")
	if err != nil {
		t.Error(err)
		return
	}
	if len(data) != 3 {
		t.Error("Expected 3 environment variables, got", len(data))
		return
	}
	if data["HOST"] != "localhost" || data["PORT"] != "123" || data["SOME_STRING"] != "this is some=string" {
		t.Error("Expected environment variables to match")
		return
	}
	if os.Getenv("HOST") != "localhost" || os.Getenv("PORT") != "123" || os.Getenv("SOME_STRING") != "this is some=string" {
		t.Error("Expected environment variables to be set correctly")
		return
	}
}
