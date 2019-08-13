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
//
// BitBox Base Config Generator
// ----------------------------
// Generates text files from template, substituting placeholders with Redis values.
// See helpText specified below for usage information.
//
// https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbconfgen
//

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gomodule/redigo/redis"
)

// Command line arguments
var (
	templateArg  = flag.String("template", "", "input template text file")
	outputArg    = flag.String("output", "", "output text file")
	redisAddrArg = flag.String("redis-addr", "localhost:6379", "redis connection address")
	redisPassArg = flag.String("redis-pass", "", "redis password")
	redisDbArg   = flag.Int("redis-db", 0, "redis database number")
	versionArg   = flag.Bool("version", false, "return program version")
	verboseArg   = flag.Bool("verbose", false, "show additional information")
	helpArg      = flag.Bool("help", false, "show help")
)

// Help text for --help option
const (
	helpText = `generates text files from a template, substituting placeholders with Redis values

Command-line arguments: 
  --template      input template text file
  --output        output text file
  --redis-addr    redis connection address (default "localhost:6379")
  --redis-db      redis database number
  --redis-pass    redis password
  --verbose
  --version
  --help

Optionally, the output file can be specified on the first line in the template text file.
This line will be dropped and only used if no --output argument is supplied.
  
  {{ #output: /tmp/output.txt }}
  
Placeholders in the template text file are defined as follows.
Make sure to respect spaces between arguments.

  {{ key }}                     is replaced by Redis 'key', only if key is present
  {{ key #rm }}                 ...deletes the placeholder if key not found
  {{ key #rmLine }}             ...deletes the whole line if key not found
  {{ key #default: some val }}  ...uses default value if key not found

`
)

func main() {
	var (
		versionNum = 0.1
		redisConn  redis.Conn
		err        error

		countReplace int
		countKeep    int
		countRm      int
		countRmLine  int
		countDefault int
	)

	flag.Parse()

	if *versionArg || *helpArg {
		fmt.Println("bbbconfgen version", versionNum)
		if *helpArg {
			fmt.Println(helpText)
		}
		os.Exit(0)
	}

	if len(*templateArg) == 0 {
		fmt.Println("No input template file specified using --template argument.")
		os.Exit(1)
	}

	// Open and check connection to Redis server
	if len(*redisPassArg) > 0 {
		redisConn, err = redis.Dial("tcp", *redisAddrArg, redis.DialDatabase(*redisDbArg))
	} else {
		redisConn, err = redis.Dial("tcp", *redisAddrArg, redis.DialPassword(*redisPassArg), redis.DialDatabase(*redisDbArg))
	}
	if err != nil {
		log.Fatal(err)
	}
	defer redisConn.Close()

	_, err = redisConn.Do("PING")
	if err != nil {
		log.Fatal(err)
	}

	// If no outputFile supplied, check first line of template file
	if len(*outputArg) == 0 {
		templateFile, err := os.Open(*templateArg)
		if err != nil {
			log.Fatal(err)
		}
		defer templateFile.Close()

		// match outputFile pattern, e.g. {{ #output: /tmp/output.txt }}
		outputFilePattern := regexp.MustCompile("{{[ ]{0,}#output: (.+?)}}")

		// read first line and extract outputFile pattern
		scannerOutputFile := bufio.NewScanner(templateFile)
		scannerOutputFile.Scan()
		firstLine := scannerOutputFile.Text()
		firstLineGroups := outputFilePattern.FindStringSubmatch(firstLine)

		// if successful, use it as *outputArg, otherwise abort
		if len(firstLineGroups) > 0 && len(firstLineGroups[1]) > 0 {
			*outputArg = strings.Trim(firstLineGroups[1], " ")
		} else {
			fmt.Println("No output file specified, specify either --output argument or within template.")
			os.Exit(1)
		}
	}

	// Open template file
	templateFile, err := os.Open(*templateArg)
	if err != nil {
		log.Fatal(err)
	}
	defer templateFile.Close()

	// Create output file
	outputFile, err := os.Create(*outputArg)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	// Read template file line by line
	scanner := bufio.NewScanner(templateFile)
	placeholderPattern := regexp.MustCompile("{{(.+?)}}")

	for scanner.Scan() {
		var (
			outputLine   string     // current line of template file that will be written to outputFile
			placeholders [][]string // contains placeholders of a single line, each with a key and optionally a fallback part
			printLine    = true     // print line if true, set to 'false' when fallback is #rmLine
		)

		outputLine = scanner.Text()
		placeholders = placeholderPattern.FindAllStringSubmatch(outputLine, -1)

		// replace individual placeholder
		for i := range placeholders {
			var (
				placeholder       string   // current placeholder
				placeholderFields []string // individual fields of placeholder, separated by a space (" ")
				redisKey          string   // Redis key of a placeholder
				redisVal          string   // value fetched from Redis for a placeholder
			)

			// Redis GET
			placeholder = placeholders[i][0]
			placeholderFields = strings.Fields(placeholders[i][1])
			redisKey = placeholderFields[0]

			// skip line with outputFile specifier
			if strings.ToLower(redisKey) == "#output:" {
				printLine = false
				break
			}

			redisVal, _ = redis.String(redisConn.Do("GET", redisKey))

			if len(redisVal) > 0 {
				// replace placeholder if Redis key is found
				outputLine = strings.Replace(outputLine, placeholder, redisVal, -1)
				countReplace++

			} else if len(placeholderFields) > 1 {
				// if specified, use fallback options if Redis key is not found
				switch strings.ToLower(placeholderFields[1]) {
				case "#rm":
					outputLine = strings.Replace(outputLine, placeholder, "", -1)
					countRm++
				case "#rmline":
					printLine = false
					countRmLine++
				case "#default:":
					defaultValue := strings.Join(placeholderFields[2:], " ")
					outputLine = strings.Replace(outputLine, placeholder, defaultValue, -1)
					countDefault++
				}
			} else {
				countKeep++
			}
		}

		// write processed line to outputFile
		if printLine {
			fmt.Fprintln(outputFile, outputLine)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if *verboseArg {
		fmt.Println("read template file", *templateArg)
		fmt.Println("written output file", *outputArg)
		fmt.Printf("%v replaced, %v kept, %v deleted, %v lines deleted, %v set to default\n\n", countReplace, countKeep, countRm, countRmLine, countDefault)
	}
}
