## What's Changed

**Full Changelog**: https://github.com/Stoufiler/caddy-jellywol/compare/{{ .PreviousTag }}...{{ .Tag }}

---

## Installation

### Docker
```bash
docker pull ghcr.io/stoufiler/caddy-jellywol:{{ .Tag }}
```

### Binary Downloads

Download the binary for your platform below. See [README](https://github.com/Stoufiler/caddy-jellywol#installation) for usage instructions.

#### Checksum Verification

```bash
# Download checksum file
wget https://github.com/Stoufiler/caddy-jellywol/releases/download/{{ .Tag }}/SHA256SUMS

# Verify (Linux/macOS example)
sha256sum -c SHA256SUMS --ignore-missing
```

#### SBOM (Software Bill of Materials)

A complete Software Bill of Materials is available for vulnerability scanning and compliance:
- [Download SBOM](https://github.com/Stoufiler/caddy-jellywol/releases/download/{{ .Tag }}/sbom.spdx.json)

---

**Note:** Compressed binaries use UPX for smaller download sizes (except macOS).
