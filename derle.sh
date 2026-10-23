#!/usr/bin/env bash
# Windows 64-bit taşınabilir exe üretir: ./derle.sh  →  Taklit.exe
set -euo pipefail
cd "$(dirname "$0")"
[ -f go.mod ] || go mod init taklit >/dev/null 2>&1
go run github.com/akavel/rsrc@v0.10.2 -manifest taklit.manifest -ico taklit.ico -arch amd64 -o rsrc_windows_amd64.syso
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go vet -unsafeptr=false ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-H windowsgui -s -w" -o Taklit.exe .
ls -la Taklit.exe
