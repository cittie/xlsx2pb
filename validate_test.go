package xlsx2pb

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileMD5(t *testing.T) {
	md5 := getFileMD5("./test/md5test")
	assert.Equal(t, "f7ffd6e04e02a743fe8bec550e64cb71", hex.EncodeToString(md5[:]))
}

func TestIsTitleValid(t *testing.T) {
	tests := []struct {
		in     string
		result bool
	}{
		{"", false},
		{"title", false},
		{"Title", true},
		{"PascalTitle", true},
		{"camelTitle", false},
		{"snake_title", false},
		//{"Snake_title", false},
	}

	for _, test := range tests {
		assert.Equal(t, test.result, IsTitleValid(test.in))
	}
}
