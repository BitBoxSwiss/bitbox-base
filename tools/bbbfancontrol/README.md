# bbbfancontrol

Simple program to control fan speed on a single board computer according to current system temperature. It's written in Go and aimed for the ROCKPro64 SBC as part of the [BitBox Base](https://github.com/digitalbitbox/bitbox-base) project by [Shift Cryptosecurity](https://shiftcrypto.ch).

The program reads the current system temperature from a single file, calculates the appropriate fan PWM value and writes it into a control file. The default values are set for the ROCKPro64 board running Armbian.

* Temperature is read from the file `/sys/class/hwmon/hwmon0/pwm1`, in °C * 1000 (e.g. `45000` for 45°C)
* Fan is controlled by writing a value between `0` (off) and `255` (max) into the file `/sys/class/thermal/thermal_zone0/temp`

The appropriate fan speed value is calculated linearly for the desired temperature range:
```
fanPWM = fanMin + ( ( fanMax - fanMin ) / ( tempMax - tempMin ) ) * ( tempCur - tempMin )
```

## Installation
The source code can be compiled directly on any single board computer 

* install Go on your SBC, e.g. the ARMv8 version: https://golang.org/dl/
* download the source code to a temporary directory named `bbbfancontrol`
* compile the source code inside the temporary directory with `go build`
* copy the resulting binary to the `/usr/local/sbin` directory 
* check by running `bbbfancontrol --help`

The program is meant to be run in the background by systemd, a sample unit file is provided. 
* copy systemd unit file `bbbfancontrol.service` into `/etc/systemd/system/` folder
* enable service with `systemctl enable bbbfancontrol`
* start service with `systemctl start bbbfancontrol` or reboot

Events are logged into journald by standard and can be viewed with
```
$ journalctl -f -u bbbfancontrol
```

## Usage
Most attributes can be supplied via command line arguments, but default values work fine for most cases.

```
$ bbbfancontrol --help

Usage of bbbfancontrol:
  -cooldown int
        temperature to cool down to in °C when stepping out of min / max temp zone (default 40)
  -cycle int
        length of sleep cycle in seconds after each temperature check (default 10)
  -fan string
        filepath to fan control file (default "/sys/class/hwmon/hwmon0/pwm1")
  -fmax int
        maximum value for fan control (default 255)
  -fmin int
        minimum value for fan control (default 120)
  -kickstart int
        seconds to kickstart fan with full power (default 0 = off)
  -temp string
        filepath to temperature value file (default "/sys/class/thermal/thermal_zone0/temp")
  -tmax int
        maximum temperature in °C for fan control, max fan above (default 60)
  -tmin int
        minimum temperature in °C for fan control, no fan below (default 45)
  -v    verbose, log internal data to stdout
  -version
        return program version
```

## Example
```
$ bbbfancontrol -v -fmin 80 -tmin 40 -tmax 55 -cycle 30 -kickstart 2

BitBox Base fan control, version 0.1
temp:      /sys/class/thermal/thermal_zone0/temp
tmin:      40
tmax:      55
cooldown:  40
fan:       /sys/class/hwmon/hwmon0/pwm1
fmin:      80
fmax:      255
kickstart: 2
cycle:     30
temperature: 39 / fan set to: 0 / kickstart: 2 / cooldown: false
Fan turned ON.
Kickstart for 2 seconds!
temperature: 45 / fan set to: 135 / kickstart: 2 / cooldown: true
temperature: 40 / fan set to: 80 / kickstart: 2 / cooldown: true
Fan turned OFF.
temperature: 39 / fan set to: 0 / kickstart: 2 / cooldown: false
...
```