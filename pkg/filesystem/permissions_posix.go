// +build !windows

package filesystem

import (
	"os"
	userpkg "os/user"
	"strconv"

	"github.com/pkg/errors"
)

type OwnershipSpecification struct {
	ownerID int

	groupID int
}

func NewOwnershipSpecification(owner, group string) (*OwnershipSpecification, error) {

	ownerID := -1
	if owner != "" {
		switch kind, identifier := ParseOwnershipIdentifier(owner); kind {
		case OwnershipIdentifierKindInvalid:
			return nil, errors.New("invalid user specification")
		case OwnershipIdentifierKindPOSIXID:
			if _, err := userpkg.LookupId(identifier); err != nil {
				return nil, errors.Wrap(err, "unable to lookup user by ID")
			} else if userID, err := strconv.Atoi(identifier); err != nil {
				return nil, errors.Wrap(err, "unable to convert user ID to numeric value")
			} else {
				ownerID = userID
			}
		case OwnershipIdentifierKindWindowsSID:
			return nil, errors.New("Windows SIDs not supported on POSIX systems")
		case OwnershipIdentifierKindName:
			if userObject, err := userpkg.Lookup(identifier); err != nil {
				return nil, errors.Wrap(err, "unable to lookup user by ID")
			} else if userID, err := strconv.Atoi(userObject.Uid); err != nil {
				return nil, errors.Wrap(err, "unable to convert user ID to numeric value")
			} else {
				ownerID = userID
			}
		default:
			panic("unhandled ownership identifier kind")
		}
	}

	groupID := -1
	if group != "" {
		switch kind, identifier := ParseOwnershipIdentifier(group); kind {
		case OwnershipIdentifierKindInvalid:
			return nil, errors.New("invalid group specification")
		case OwnershipIdentifierKindPOSIXID:
			if _, err := userpkg.LookupGroupId(identifier); err != nil {
				return nil, errors.Wrap(err, "unable to lookup group by ID")
			} else if g, err := strconv.Atoi(identifier); err != nil {
				return nil, errors.Wrap(err, "unable to convert group ID to numeric value")
			} else {
				groupID = g
			}
		case OwnershipIdentifierKindWindowsSID:
			return nil, errors.New("Windows SIDs not supported on POSIX systems")
		case OwnershipIdentifierKindName:
			if groupObject, err := userpkg.LookupGroup(identifier); err != nil {
				return nil, errors.Wrap(err, "unable to lookup group by ID")
			} else if g, err := strconv.Atoi(groupObject.Gid); err != nil {
				return nil, errors.Wrap(err, "unable to convert group ID to numeric value")
			} else {
				groupID = g
			}
		default:
			panic("unhandled ownership identifier kind")
		}
	}

	return &OwnershipSpecification{
		ownerID: ownerID,
		groupID: groupID,
	}, nil
}

func SetPermissionsByPath(path string, ownership *OwnershipSpecification, mode Mode) error {

	if ownership != nil && (ownership.ownerID != -1 || ownership.groupID != -1) {
		if err := os.Chown(path, ownership.ownerID, ownership.groupID); err != nil {
			return errors.Wrap(err, "unable to set ownership information")
		}
	}

	mode = mode & ModePermissionsMask
	if mode != 0 {
		if err := os.Chmod(path, os.FileMode(mode)); err != nil {
			return errors.Wrap(err, "unable to set permission bits")
		}
	}

	return nil
}
