#! /bin/bash
set -euo pipefail

go build -o verbatim
sudo setcap cap_net_bind_service=ep ./verbatim
./verbatim
