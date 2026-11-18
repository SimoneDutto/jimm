#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/../.."
ca=$(microjaas certificate)
printf '%s\n' "$ca" | openssl x509 -noout > /dev/null
{
    cat <<'EOF'
cloudinit-userdata: |
  preruncmd:
    - echo "192.168.1.14    microjaas.test" >> /etc/hosts
  ca-certs:
    trusted:
      - |
EOF
    printf '%s\n' "$ca" | sed 's/^/        /'
} > local/microjaas-demo/cloudinit.yaml
