package filesystem

import (
	"os"
	userpkg "os/user"

	"github.com/pkg/errors"

	"golang.org/x/sys/windows"

	aclapi "github.com/hectane/go-acl/api"
)

type OwnershipSpecification struct {
	ownerSID *windows.SID

	groupSID *windows.SID
}

func NewOwnershipSpecification(owner, group string) (*OwnershipSpecification, error) {

	var ownerSID *windows.SID
	if owner != "" {
		switch kind, identifier := ParseOwnershipIdentifier(owner); kind {
		case OwnershipIdentifierKindInvalid:
			return nil, errors.New("invalid owner specification")
		case OwnershipIdentifierKindPOSIXID:
			return nil, errors.New("POSIX IDs not supported on Windows systems")
		case OwnershipIdentifierKindWindowsSID:

			var retrievedSID string
			if userObject, err := userpkg.LookupId(identifier); err == nil {
				retrievedSID = userObject.Uid
			} else if groupObject, err := userpkg.LookupGroupId(identifier); err == nil {
				retrievedSID = groupObject.Gid
			} else {
				return nil, errors.New("unable to find user or group with specified owner SID")
			}

			if s, err := windows.StringToSid(retrievedSID); err != nil {
				return nil, errors.Wrap(err, "unable to convert SID string to object")
			} else {
				ownerSID = s
			}
		case OwnershipIdentifierKindName:
			var retrievedSID string
			if userObject, err := userpkg.Lookup(identifier); err == nil {
				retrievedSID = userObject.Uid
			} else if groupObject, err := userpkg.LookupGroup(identifier); err == nil {
				retrievedSID = groupObject.Gid
			} else {
				return nil, errors.New("unable to find user or group with specified owner name")
			}

			if s, err := windows.StringToSid(retrievedSID); err != nil {
				return nil, errors.Wrap(err, "unable to convert SID string to object")
			} else {
				ownerSID = s
			}
		default:
			panic("unhandled ownership identifier kind")
		}
	}

	var groupSID *windows.SID
	if group != "" {
		switch kind, identifier := ParseOwnershipIdentifier(group); kind {
		case OwnershipIdentifierKindInvalid:
			return nil, errors.New("invalid group specification")
		case OwnershipIdentifierKindPOSIXID:
			return nil, errors.New("POSIX IDs not supported on Windows systems")
		case OwnershipIdentifierKindWindowsSID:
			if groupObject, err := userpkg.LookupGroupId(identifier); err != nil {
				return nil, errors.Wrap(err, "unable to lookup group by ID")
			} else if g, err := windows.StringToSid(groupObject.Gid); err != nil {
				return nil, errors.Wrap(err, "unable to convert SID string to object")
			} else {
				groupSID = g
			}
		case OwnershipIdentifierKindName:
			if groupObject, err := userpkg.LookupGroup(identifier); err != nil {
				return nil, errors.Wrap(err, "unable to lookup group by ID")
			} else if g, err := windows.StringToSid(groupObject.Gid); err != nil {
				return nil, errors.Wrap(err, "unable to convert SID string to object")
			} else {
				groupSID = g
			}
		default:
			panic("unhandled ownership identifier kind")
		}
	}

	return &OwnershipSpecification{
		ownerSID: ownerSID,
		groupSID: groupSID,
	}, nil
}

func SetPermissionsByPath(path string, ownership *OwnershipSpecification, mode Mode) error {
	if ownership != nil && (ownership.ownerSID != nil || ownership.groupSID != nil) {
		var information uint32
		if ownership.ownerSID != nil {
			information |= aclapi.OWNER_SECURITY_INFORMATION
		}
		if ownership.groupSID != nil {
			information |= aclapi.GROUP_SECURITY_INFORMATION
		}

		if err := aclapi.SetNamedSecurityInfo(
			path,
			aclapi.SE_FILE_OBJECT,
			information,
			ownership.ownerSID,
			ownership.groupSID,
			0,
			0,
		); err != nil {
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
