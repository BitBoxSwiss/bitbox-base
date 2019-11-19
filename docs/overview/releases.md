---
layout: default
title: Releases
parent: Overview
nav_order: 160
---
## Releases

The BitBoxBase releases are tagged frequently on GitHub so that specific versions can be built from source. They are also released as binary files on the [GitHub Releases](https://github.com/digitalbitbox/bitbox-base/releases/) page.

### Image types

Releases come in two flaviours:

* Over-the-air (OTA) update: the BitBoxApp will notify you if a new update is available. If you choose to install it, the BitBoxBase automatically downloads a signed update artefact (`.base`) and applies the update as explained in [Updates / Process](../update/update-process.md). This update artefact can also be applied manually from the command line.

* Full disk image: the `.tar.gz` archive contains a disk image with all necessary partitions that can be written directly to an eMMC chip.

### Verifying the release binaries

The OTA artefacts are signed by the Shift BitBoxBase release key, which is checked against the verification key that is already present in the BitBoxBase.

When updating manually, either allowing unsigned OTA artefacts or flashing the full disk image, you need to verify the release binaries yourself to avoid using an inofficial and potentially malicious image. You can verify the checksums of all released binaries against the file SHA256SUMS.asc that is signed by Stadicus.

1. Get Stadicus' public PGP key:

```sh
gpg --keyserver hkp://keyserver.ubuntu.com --receive-keys 82AB582358C37100221A0FA8CF4D0ACF957AF4AD
```

2. Verify signature of `SHA256SUMS.asc`:

```sh
$ gpg --verify SHA256SUMS.asc

gpg: Signature made So 22 Sep 2019 16:55:18 CEST
gpg:                using RSA key 863CF135BDC28B36AB902CCA0B66622A0EB6951B
gpg: Good signature from "Stadicus <stadicus@protonmail.com>"
Primary key fingerprint: 82AB 5823 58C3 7100 221A  0FA8 CF4D 0ACF 957A F4AD
     Subkey fingerprint: 863C F135 BDC2 8B36 AB90  2CCA 0B66 622A 0EB6 951B
```

3. Verify the SHA256 checksums of the binary release files:

```sh
$ sha256sum --check SHA256SUMS.asc --ignore-missing

BitBoxBase-v0.X.X-RockPro64.base: OK
BitBoxBase-v0.X.X-RockPro64.tar.gz: OK
sha256sum: WARNING: 19 lines are improperly formatted
```
