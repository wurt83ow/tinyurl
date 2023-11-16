package shorturl

import (
	"testing"
)

func TestShorten(t *testing.T) {

	type test struct {
		name, url, shortURLAdress, key, shurl string
	}

	test1 := test{
		name:           "simple test #1",
		url:            "https://practicum.yandex.ru/",
		shortURLAdress: "http://localhost:8080/",
		key:            "nOykhckC3Od",
		shurl:          "http://localhost:8080/nOykhckC3Od",
	}

	tests := []test{
		test1,
	}
	test2 := test1
	test2.shortURLAdress = test2.shortURLAdress[:len(test2.shortURLAdress)-1]
	tests = append(tests, test2)

	for _, test := range tests { // cycle through all tests
		t.Run(test.name, func(t *testing.T) {
			if key, shurl := Shorten(test.url, test.shortURLAdress); key != test.key || shurl != test.shurl {
				t.Errorf(shurl)
			}
		})
	}
}
