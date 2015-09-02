apport-core-unpacker
====================

# Overview
Use this script to extract core dumps from Ubuntu problem report files.
I wrote it since the stock apport-unpack could not handle 3.9Gb worth of core.

# Usage
    $ ./apport-core-unpacker -path=<path-to-problem-report>

It outputs a CoreDump.core file in the current directory.
