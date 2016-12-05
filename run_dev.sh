#! /bin/bash
set -euo pipefail
IFS=$'\n\t'

go build -o verbatim
sudo setcap cap_net_bind_service=ep ./verbatim
./verbatim
