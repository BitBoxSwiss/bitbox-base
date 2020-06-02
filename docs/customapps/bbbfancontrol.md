---
layout: default
title: Fancontrol
nav_order: 130
parent: Custom applications
---
## BitBoxBase Fancontrol

Simple program to control fan speed on a single board computer according to current system temperature.
It's written in Go and aimed for the ROCKPro64 SBC as part of the [BitBoxBase](https://github.com/digitalbitbox/bitbox-base) project.

The program reads the current system temperature from a single file, calculates the appropriate fan PWM value and writes it into a control file.
The default values are set for the ROCKPro64 board running Armbian.

* Temperature is read from the file `/sys/class/hwmon/hwmon0/pwm1`, in °C * 1000 (e.g. `45000` for 45°C)
* Fan is controlled by writing a value between `0` (off) and `255` (max) into the file `/sys/class/thermal/thermal_zone0/temp`

[See Docs on GitHub](https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbfancontrol){: .btn }
