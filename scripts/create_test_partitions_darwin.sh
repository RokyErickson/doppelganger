#!/bin/bash

hdiutil create -megabytes 50 -fs "MS-DOS FAT32" -volname FAT32ROOT -o fat32image.dmg || exit $?
hdiutil attach fat32image.dmg || exit $?
export DOPPELGANGER_TEST_FAT32_ROOT="/Volumes/FAT32ROOT"

hdiutil create -megabytes 50 -fs "HFS+" -volname "HFSRoot" -o hfsimage.dmg || exit $?
hdiutil attach hfsimage.dmg || exit $?
export DOPPELGANGER_TEST_HFS_ROOT="/Volumes/HFSRoot"

hdiutil create -megabytes 50 -fs "APFS" -volname "APFSRoot" -o apfsimage.dmg || exit $?
hdiutil attach apfsimage.dmg || exit $?
export DOPPELGANGER_TEST_APFS_ROOT="/Volumes/APFSRoot"

hdiutil create -megabytes 50 -fs "MS-DOS FAT32" -volname FAT32SUB -o fat32subimage.dmg || exit $?
hdiutil attach -mountroot "${DOPPELGANGER_TEST_APFS_ROOT}" fat32subimage.dmg || exit $?
export DOPPELGANGER_TEST_FAT32_SUBROOT="${DOPPELGANGER_TEST_APFS_ROOT}/FAT32SUB"
