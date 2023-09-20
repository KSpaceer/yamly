package generator

import (
	"reflect"
	"strings"
)

type fieldTags struct {
	name string

	omitField bool
	omitempty bool
	inline    bool
}

func parseTags(f reflect.StructTag) fieldTags {
	t := fieldTags{}

	options := strings.Split(f.Get("yaml"), ",")

	if len(options) == 1 && options[0] == "-" {
		t.omitField = true
	}

	for i, s := range options {
		switch {
		case i == 0:
			t.name = s
		case s == "omitempty":
			t.omitempty = true
		case s == "inline":
			t.inline = true
		}
	}

	return t
}
