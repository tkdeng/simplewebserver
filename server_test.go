package server

import "testing"

func Test(t *testing.T) {
	app, err := New("./test")
	if err != nil {
		t.Error(err)
	}

	app.Listen()
}
