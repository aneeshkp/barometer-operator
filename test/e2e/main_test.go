package e2e

import (
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
)

type testArgs struct {
	e2eImageName *string
}

var args = &testArgs{}

//https://github.com/operator-framework/operator-sdk/blob/2f772d1dc2340dd19bdc3ec8c2dc9f0f77cc8297/doc/test-framework/writing-e2e-tests.md
func TestMain(m *testing.M) {
	framework.MainEntry(m)
}
