#!/usr/bin/env bash
set -euo pipefail

git describe --tags --abbrev=0 > version.txt
