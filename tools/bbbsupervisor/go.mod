module github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor

go 1.13

require (
	github.com/digitalbitbox/bitbox-base/middleware v0.0.0-20191129152738-f894e6d1448b
	github.com/digitalbitbox/bitbox02-api-go v0.0.0-20191122093321-5bacb3c08094s
	github.com/tidwall/gjson v1.3.4
)

replace github.com/digitalbitbox/bitbox02-api-go => /home/b10c/shift/bitbox02-api-go // TODO: remove before merge
