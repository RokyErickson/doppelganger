#!/bin/bash

set -e

DOPPELGANGER_VERSION=$(build/doppelganger version)

tar xzf build/release/doppelganger_windows_386_v${DOPPELGANGER_VERSION}.tar.gz
zip build/release/doppelganger_windows_386_v${DOPPELGANGER_VERSION}.zip doppelganger.exe doppelganger-agents.tar.gz
rm doppelganger.exe doppelganger-agents.tar.gz

tar xzf build/release/doppelganger_windows_amd64_v${DOPPELGANGER_VERSION}.tar.gz
zip build/release/doppelganger_windows_amd64_v${DOPPELGANGER_VERSION}.zip doppelganger.exe doppelganger-agents.tar.gz
rm doppelganger.exe doppelganger-agents.tar.gz
