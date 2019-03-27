package filesystem

import (
	"testing"
)

func TestModePermissionMaskIsExpected(t *testing.T) {
	if ModePermissionsMask != Mode(0777) {
		t.Error("ModePermissionsMask value not equal to expected:", ModePermissionsMask, "!=", Mode(0777))
	}
}

func TestModePermissionMaskIsUnionOfPermissions(t *testing.T) {
	permissionUnion := ModePermissionUserRead | ModePermissionUserWrite | ModePermissionUserExecute |
		ModePermissionGroupRead | ModePermissionGroupWrite | ModePermissionGroupExecute |
		ModePermissionOthersRead | ModePermissionOthersWrite | ModePermissionOthersExecute
	if ModePermissionsMask != permissionUnion {
		t.Error("ModePermissionsMask value not equal to union of permissions:", ModePermissionsMask, "!=", permissionUnion)
	}
}

type parseModeTestCase struct {
	value string

	mask Mode

	expectFailure bool

	expected Mode
}

func (c *parseModeTestCase) run(t *testing.T) {

	t.Helper()

	if result, err := ParseMode(c.value, c.mask); err == nil && c.expectFailure {
		t.Fatal("parsing succeeded when failure was expected")
	} else if err != nil && !c.expectFailure {
		t.Fatal("parsing failed unexpectedly:", err)
	} else if result != c.expected {
		t.Error("parsing result does not match expected:", result, "!=", c.expected)
	}
}

func TestParseModeEmpty(t *testing.T) {

	testCase := &parseModeTestCase{
		mask:          ModePermissionsMask,
		expectFailure: true,
	}

	testCase.run(t)
}

func TestParseModeInvalid(t *testing.T) {

	testCase := &parseModeTestCase{
		value:         "laksjfd",
		mask:          ModePermissionsMask,
		expectFailure: true,
	}

	testCase.run(t)
}

func TestParseModeOverflow(t *testing.T) {

	testCase := &parseModeTestCase{
		value:         "45201371000",
		mask:          ModePermissionsMask,
		expectFailure: true,
	}

	testCase.run(t)
}

func TestParseModeInvalidBits(t *testing.T) {

	testCase := &parseModeTestCase{
		value:         "1000",
		mask:          ModePermissionsMask,
		expectFailure: true,
	}

	testCase.run(t)
}

func TestParseModeValid(t *testing.T) {

	testCase := &parseModeTestCase{
		value:    "777",
		mask:     ModePermissionsMask,
		expected: 0777,
	}

	testCase.run(t)
}

func TestParseModeValidWithZeroPrefix(t *testing.T) {

	testCase := &parseModeTestCase{
		value:    "0755",
		mask:     ModePermissionsMask,
		expected: 0755,
	}

	testCase.run(t)
}

func TestParseModeValidWithMultiZeroPrefix(t *testing.T) {

	testCase := &parseModeTestCase{
		value:    "00644",
		mask:     ModePermissionsMask,
		expected: 0644,
	}

	testCase.run(t)
}
