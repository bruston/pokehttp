package main

import "testing"

func TestGetTitle(t *testing.T) {
	tt := []struct {
		in       []byte
		expected string
	}{
		{
			[]byte("ASDF <title>Hello, world!</title>"),
			"Hello, world!",
		},
		{
			[]byte("ASDF<tiTLE  >I am a title </tiTLE  >"),
			"I am a title",
		},
		{
			[]byte("ASDF <title>I am\na title</title>\nASDF"),
			"I am a title",
		},
		{
			[]byte("ASDF ASDF<title >Title Here </title   >\n\n<title asd>asdf</title>"),
			"Title Here",
		},
	}
	for i, v := range tt {
		title := getTitle(v.in)
		if title != v.expected {
			t.Errorf("%d. expecting %s, got %s for input %s\n", i, v.expected, title, v.in)
		}
	}
}
