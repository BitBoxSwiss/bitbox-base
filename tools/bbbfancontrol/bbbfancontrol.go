// Copyright 2019 Shift Cryptosecurity AG, Switzerland.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func readValueFile(filepath string) (value string, err error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	value = string(data)
	return value, err
}

func writeValueFile(filepath string, value string) (err error) {
	output := []byte(value)
	err = ioutil.WriteFile(filepath, output, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func main() {

	versionNum := 0.1

	// parse command line arguments
	tempFile := flag.String("temp", "/sys/class/thermal/thermal_zone0/temp", "filepath to temperature value file")
	tempMin := flag.Int("tmin", 45, "minimum temperature in °C for fan control, no fan below")
	tempMax := flag.Int("tmax", 60, "maximum temperature in °C for fan control, max fan above")
	tempCooldown := flag.Int("cooldown", 40, "temperature to cool down to in °C when stepping out of min / max temp zone")
	fanFile := flag.String("fan", "/sys/class/hwmon/hwmon0/pwm1", "filepath to fan control file")
	fanMin := flag.Int("fmin", 120, "minimum value for fan control")
	fanMax := flag.Int("fmax", 255, "maximum value for fan control")
	fanKickstart := flag.Int("kickstart", 0, "seconds to kickstart fan with full power (default 0 = off)")
	cycle := flag.Int("cycle", 10, "length of sleep cycle in seconds after each temperature check")
	verbose := flag.Bool("v", false, "verbose, log internal data to stdout")
	version := flag.Bool("version", false, "return program version")
	flag.Parse()

	if *version {
		fmt.Println("bbbfancontrol version", versionNum)
		os.Exit(0)
	}

	// sanity check for arguments
	if *tempCooldown > *tempMin || *tempMin > *tempMax {
		fmt.Printf("ERROR: inconsistent temperature range supplied.\n")
		fmt.Printf("       cooldown (%v) must be <= tmin (%v) must be < tmax (%v)\n\n", *tempCooldown, *tempMin, *tempMax)
		os.Exit(1)
	}

	fmt.Println("BitBox Base fan control, version", versionNum)
	if *verbose {
		fmt.Println("temp:     ", *tempFile)
		fmt.Println("tmin:     ", *tempMin)
		fmt.Println("tmax:     ", *tempMax)
		fmt.Println("cooldown: ", *tempCooldown)
		fmt.Println("fan:      ", *fanFile)
		fmt.Println("fmin:     ", *fanMin)
		fmt.Println("fmax:     ", *fanMax)
		fmt.Println("kickstart:", *fanKickstart)
		fmt.Println("cycle:    ", *cycle)
	}

	cooldown := false
	for {
		// read current temperature
		tempStr, _ := readValueFile(*tempFile)
		tempCur, err := strconv.Atoi(strings.TrimSpace(tempStr))
		tempCur = tempCur / 1000
		if err != nil {
			log.Fatal(err)
		}

		// linear PWM increase beteween tempMin and tempMax, from fanMin to fanMax
		fanPWM := *fanMin + ((*fanMax-*fanMin)/(*tempMax-*tempMin))*(tempCur-*tempMin)

		// upper cap fanPWM to fanMax
		if fanPWM > *fanMax {
			fanPWM = *fanMax
		}

		if cooldown {
			// keep fanPWM at fanMin during cooldown
			if fanPWM < *fanMin {
				fanPWM = *fanMin
			}

			// cooldown period finished, fan off
			if tempCur < *tempCooldown {
				fanPWM = 0
				cooldown = false
				fmt.Println("Fan turned OFF.")
			}

		} else if tempCur >= *tempMin {
			// tempMin exeeded, start fan
			cooldown = true
			fmt.Println("Fan turned ON.")

			if *fanKickstart > 0 {
				if *verbose {
					fmt.Printf("Kickstart for %v seconds!\n", *fanKickstart)
				}
				writeValueFile(*fanFile, strconv.Itoa(*fanMax))
				time.Sleep(time.Duration(*fanKickstart) * time.Second)
			}
		} else {
			fanPWM = 0
		}
		// adjust fan speed
		writeValueFile(*fanFile, strconv.Itoa(fanPWM))

		if *verbose {
			fmt.Printf("temperature: %v / fan set to: %v / kickstart: %v / cooldown: %v\n", tempCur, fanPWM, *fanKickstart, cooldown)
		}

		time.Sleep(time.Duration(*cycle) * time.Second)
	}
}
