#! /bin/bash
set -euo pipefail
IFS=$'\n\t'

go build -o verbatim
if [[ "$OSTYPE" == "linux-gnu" ]]; then
  sudo setcap cap_net_bind_service=ep ./verbatim
fi
./verbatim
