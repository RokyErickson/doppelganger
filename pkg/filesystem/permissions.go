package filesystem

import (
	"strings"
)

type OwnershipIdentifierKind uint8

const (
	OwnershipIdentifierKindInvalid OwnershipIdentifierKind = iota

	OwnershipIdentifierKindPOSIXID

	OwnershipIdentifierKindWindowsSID

	OwnershipIdentifierKindName
)

func isValidPOSIXID(value string) bool {

	if len(value) == 0 {
		return false
	}

	if value == "0" {
		return true
	}

	first := true
	for _, r := range value {
		if first {
			if !('1' <= r && r <= '9') {
				return false
			}
			first = false
		} else {
			if !('0' <= r && r <= '9') {
				return false
			}
		}
	}

	return true
}

func isValidWindowsSID(value string) bool {

	if len(value) == 0 {
		return false
	}

	return true
}

func ParseOwnershipIdentifier(specification string) (OwnershipIdentifierKind, string) {

	if len(specification) == 0 {
		return OwnershipIdentifierKindInvalid, ""
	}

	if strings.HasPrefix(specification, "id:") {
		if value := specification[3:]; !isValidPOSIXID(value) {
			return OwnershipIdentifierKindInvalid, ""
		} else {
			return OwnershipIdentifierKindPOSIXID, value
		}
	}

	if strings.HasPrefix(specification, "sid:") {
		if value := specification[4:]; !isValidWindowsSID(value) {
			return OwnershipIdentifierKindInvalid, ""
		} else {
			return OwnershipIdentifierKindWindowsSID, value
		}
	}

	return OwnershipIdentifierKindName, specification
}
