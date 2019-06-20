---
layout: default
title: Platform choice
parent: Hardware
nav_order: 220
---

## Platform Choice

Various hardware platforms were considered for the BitBox Base. The most notable being the Odroid HC1, Rock64 and RockPro64. When choosing our platform, the main metrics we used were performance (blockchain sync. speed), expandability and form factor.

### Performance

Considering the metrics above, it was found that the RockPro64 is an extremely viable choice for running a Bitcoin full node. The Rockchip RK3399 is an efficient yet powerful CPU due to its 64-bit hex-core single board computer (SBC), which includes 4x 1.5GHz cores and 2x high performance 2x 2.0GHz cores.
This allows for fast sync. times (approximately 1.5 days over Tor) while remaining efficient when synching is complete. Moreover, the fast LPDDR4 4GB RAM allows for a fast initial block download (IBD) time and greater dbcache allocation (2000MB).
In addition, it allows for high-performance services to run, such as NGINX, Prometheus and Grafana, without significant limitations.

### Form factor

With regards to form factor, the RockPro64 has external ports conveniently placed for a plug-n-play node while still being very compact; approximately 90 x 135 mm. The power and ethernet ports are at the back and power button/restart buttons at the front.

### Expandibility and Interfaces

The RockPro64 has many external/internal interfaces which allow for future expandability. These interfaces include the ability to add future internal modules such as wifi/bluetooth and/or a camera.
Furthermore, the internal PCIe port allows for fast internal storage. In addition, the PI-2 connecter allows for even more expandability and will be used to connect our custom HSM to the RockPro64.
With regards to external interfaces, the RockPro64 has HDMI-out, 2x USB 2.0, 1x USB 1.0 and 1x USB-C.

### Future Support

Furthermore, Pine64 commits to supply and support the RockPro64 for at least 5 years (2023) and it is likely they will continue to do so for longer.
For more information about the RockPro64, please visit their [Wiki](https://wiki.pine64.org/index.php/ROCKPro64_Main_Page).
