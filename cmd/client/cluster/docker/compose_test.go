package docker

import (
	"testing"
)

func TestNewDockerCompose(t *testing.T) {

	compose := NewDockerCompose(nil)
	err := compose.Up()
	if err != nil {
		t.Fatal(err)
	}
	compose.PrintInfo()
	defer compose.Down()

	err = compose.StopValidator(0)
	if err != nil {
		t.Fatal(err)
	}
	err = compose.StartValidator(0)
	if err != nil {
		t.Fatal(err)
	}
}
