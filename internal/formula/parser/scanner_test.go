package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindAllIdentifiers(t *testing.T) {
	identifiers := FindAllIdentifiers("var1 + var2 * 10 / 2 - var3")
	assert.Equal(t, []string{"var1", "var2", "var3"}, identifiers)
}
