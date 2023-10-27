package llmi18n

import (
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestTranslatei18nYAMLFile(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipping test in CI environment")
	}
	c := qt.New(t)
	b, err := os.ReadFile("testdata/en.yaml")
	c.Assert(err, qt.IsNil)

	res, err := TranslatedQuotedStrings(string(b))
	c.Assert(err, qt.IsNil)
	c.Assert(res, qt.Contains, "one: \"Übersetzung\"")
}
