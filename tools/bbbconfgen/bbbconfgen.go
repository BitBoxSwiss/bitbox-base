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
// Generates configuration files from template, substituting placeholders with Redis values.
// See helpText specified below for usage information.
//
// https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbconfgen
//

package main

import (
	"bufio"
	"errors"
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
	templateArg  = flag.String("template", "", "input template config file")
	outputArg    = flag.String("output", "", "output config file")
	redisAddrArg = flag.String("redis-addr", "localhost:6379", "redis connection address")
	redisPassArg = flag.String("redis-pass", "", "redis password")
	redisDbArg   = flag.Int("redis-db", 0, "redis database number")
	versionArg   = flag.Bool("version", false, "return program version")
	quietArg     = flag.Bool("quiet", false, "suppress parsing information")
	helpArg      = flag.Bool("help", false, "show help")
)

// Help text for --help option
const (
	helpText = `generates configuration files from a template, substituting placeholders with Redis values

Command-line arguments: 
  --template      input template config file
  --output        output config file
  --redis-addr    redis connection address  (default "localhost:6379")
  --redis-db      redis database number     (default 0)
  --redis-pass    redis password
  --version
  --quiet
  --help

Optionally, the output file can be specified on the first line in the template file.
This line will be dropped and only used if no --output argument is supplied.
  
  {{ #output: /tmp/output.conf }}
  
Placeholders in the template file are defined as follows.
Make sure to respect spaces between arguments.
  
  {{ key }}                     is replaced by Redis 'key', only if key is present
  {{ key #rm }}                 ...deletes the placeholder if key not found
  {{ key #rmLine }}             ...deletes the whole line if key not found
  {{ key #default: some val }}  ...uses default value if key not found

`
)

func initialize(versionNum float64) {
	// parse command line arguments
	flag.Parse()

	// remove timestamp from logger
	log.SetFlags(0)

	if *versionArg || *helpArg {
		log.Println("bbbconfgen version", versionNum)
		if *helpArg {
			fmt.Println(helpText)
		}
		os.Exit(0)
	}

	if len(*templateArg) == 0 {
		log.Fatalln("No input template file specified using --template argument.")
	}
}

// Open and check connection to Redis server
func connectRedis() (r redis.Conn, err error) {
	if len(*redisPassArg) > 0 {
		r, err = redis.Dial("tcp", *redisAddrArg, redis.DialDatabase(*redisDbArg))
	} else {
		r, err = redis.Dial("tcp", *redisAddrArg, redis.DialPassword(*redisPassArg), redis.DialDatabase(*redisDbArg))
	}
	if err != nil {
		return nil, err
	}

	_, err = r.Do("PING")
	return r, err
}

// open template configuration file
func openTemplateFile() (fp *os.File, err error) {
	fp, err = os.Open(*templateArg)
	return
}

// open output file, path provided either by cli or read it from template file
func openOutputFile() (filepointer *os.File, filename string, err error) {

	if len(*outputArg) > 0 {
		filename = *outputArg

	} else {
		// if no cli outputFile provided,
		templateFile, err := os.Open(*templateArg)
		if err != nil {
			return nil, "", errors.New("cannot open templateFile " + *templateArg)
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
			filename = strings.Trim(firstLineGroups[1], " ")
		} else {
			return nil, "", errors.New("no output file specified, specify either --output argument or within template")
		}
	}

	filepointer, err = os.Create(filename)
	if err != nil {
		return nil, "", errors.New("cannot create outputFile " + filename)
	}
	return
}

// parse template config file and replace placeholders with Redis values
// also, count number of replacements
func parseTemplate(redisConn redis.Conn, templateFile *os.File, outputFile *os.File) (err error) {

	var (
		countLines   int
		countReplace int
		countKeep    int
		countRm      int
		countRmLine  int
		countDefault int
	)

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
			countLines++
			if err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !*quietArg {
		log.Printf("written %v lines\n", countLines)
		log.Printf("placeholders: %v replaced, %v kept, %v deleted, %v lines deleted, %v set to default\n\n", countReplace, countKeep, countRm, countRmLine, countDefault)
	}
	return nil

}

func main() {
	var (
		versionNum = 0.1
		redisConn  redis.Conn
		err        error
	)

	// parse cli arguments with some sanity checks
	initialize(versionNum)

	// connect to Redis
	redisConn, err = connectRedis()
	if err != nil {
		log.Fatal(err)
	}
	defer redisConn.Close()
	if !*quietArg {
		log.Println("connected to Redis")
	}

	// open template file
	templateFile, err := openTemplateFile()
	if err != nil {
		log.Fatal(err)
	}
	defer templateFile.Close()
	if !*quietArg {
		log.Println("opened template config file", *templateArg)
	}

	// open outputFile, either from cli or from template file
	outputFile, outputFilename, err := openOutputFile()
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()
	if !*quietArg {
		log.Println("writing into output file", outputFilename)
	}

	// parse temlateFile
	err = parseTemplate(redisConn, templateFile, outputFile)
	if err != nil {
		log.Fatal(err)
	}

}
