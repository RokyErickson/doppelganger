package filesystem

import (
	"testing"
)

type parseOwnershipIdentifierTestCase struct {
	specification string

	expectedKind OwnershipIdentifierKind

	expectedValue string
}

func (c *parseOwnershipIdentifierTestCase) run(t *testing.T) {

	t.Helper()

	kind, value := ParseOwnershipIdentifier(c.specification)

	if kind != c.expectedKind {
		t.Error("parsed kind does not match expected:", kind, "!=", c.expectedKind)
	}
	if value != c.expectedValue {
		t.Error("parsed value does not match expected:", value, "!=", c.expectedValue)
	}
}

func TestParseOwnershipIdentifierEmpty(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "",
		expectedKind:  OwnershipIdentifierKindInvalid,
		expectedValue: "",
	}

	testCase.run(t)
}

func TestParseOwnershipIdentifierPOSIXIDEmpty(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "id:",
		expectedKind:  OwnershipIdentifierKindInvalid,
		expectedValue: "",
	}

	testCase.run(t)
}

func TestParseOwnershipIdentifierPOSIXIDOctal(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "id:0442",
		expectedKind:  OwnershipIdentifierKindInvalid,
		expectedValue: "",
	}

	testCase.run(t)
}

func TestParseOwnershipIdentifierPOSIXIDAlphaNumeric(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "id:5a42",
		expectedKind:  OwnershipIdentifierKindInvalid,
		expectedValue: "",
	}

	testCase.run(t)
}

func TestParseOwnershipIdentifierPOSIXIDRoot(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "id:0",
		expectedKind:  OwnershipIdentifierKindPOSIXID,
		expectedValue: "0",
	}

	testCase.run(t)
}

func TestParseOwnershipIdentifierPOSIXIDSingleDigit(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "id:4",
		expectedKind:  OwnershipIdentifierKindPOSIXID,
		expectedValue: "4",
	}

	testCase.run(t)
}

func TestParseOwnershipIdentifierPOSIXIDMultiDigit(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "id:454",
		expectedKind:  OwnershipIdentifierKindPOSIXID,
		expectedValue: "454",
	}

	testCase.run(t)
}

func TestParseOwnershipIdentifierWindowsSIDEmpty(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "sid:",
		expectedKind:  OwnershipIdentifierKindInvalid,
		expectedValue: "",
	}

	testCase.run(t)
}

func TestParseOwnershipIdentifierWindowsSIDStringConstant(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "sid:BA",
		expectedKind:  OwnershipIdentifierKindWindowsSID,
		expectedValue: "BA",
	}

	testCase.run(t)
}

func TestParseOwnershipIdentifierWindowsSIDWellKnown(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "sid:S-1-3-0",
		expectedKind:  OwnershipIdentifierKindWindowsSID,
		expectedValue: "S-1-3-0",
	}

	testCase.run(t)
}

func TestParseOwnershipIdentifierName(t *testing.T) {

	testCase := &parseOwnershipIdentifierTestCase{
		specification: "george",
		expectedKind:  OwnershipIdentifierKindName,
		expectedValue: "george",
	}

	testCase.run(t)
}
