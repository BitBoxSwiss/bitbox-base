# bbbconfgen

Application to generate text files from a template, replacing placeholders that specify Redis keys.
It's written in Go as part of the [BitBox Base](https://github.com/digitalbitbox/bitbox-base) project by [Shift Cryptosecurity](https://shiftcrypto.ch) and used for automatically generating configuration files.

The program reads a text file specified using the `--template` argument, parses the contents and writes it into a target file that is either specified with the `--output` argument, or directly on the first line of the template.

## Usage

The program is run from the command line.

```console
$ bbbconfgen --help

bbbconfgen version 0.1
generates text files from a template, substituting placeholders with Redis values

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
  
  {{ #output: /tmp/output.conf }}
  
Placeholders in the template text file are defined as follows.
Make sure to respect spaces between arguments.

  {{ key }}                     is replaced by Redis 'key', only if key is present
  {{ key #rm }}                 ...deletes the placeholder if key not found
  {{ key #rmLine }}             ...deletes the whole line if key not found
  {{ key #default: some val }}  ...uses default value if key not found
```

## Example

### Prerequisites

The environment must provide a running Redis server.

### Import example Redis values

To prepare the example data, import some keys into Redis and check them.

```bash
$ redis-cli SET bitcoind:mainnet 1
OK
$ redis-cli SET bitcoind:rpcconnect 127.0.0.1
OK
$ redis-cli SET bitcoind:seednode:1 nkf5e6b7pl4jfd4a.onion
OK

$ redis-cli KEYS 'bitcoin*'
1) "bitcoind:seednode:1"
2) "bitcoind:rpcconnect"
3) "bitcoind:mainnet"
```

### Creating text file from template

Create the new file `bitcoin-output.conf` based on the template and the Redis key/value pairs.

```console
$ cat test/bitcoin-template.conf
# network
mainnet={{ bitcoind:mainnet }}
testnet={{ bitcoind:testnet #default: 0 }}
rpcconnect={{ bitcoind:rpcconnect }}
dbcache={{ bitcoind:dbcache #default: 300 }}
seednode={{ bitcoind:seednode:1 #rmLine }}
seednode={{ bitcoind:seednode:2 #rmLine }}
seednode={{ bitcoind:seednode:3 #rmLine }}

$ ./bbbconfgen --template test/bitcoin-template.conf --output test/bitcoin-output.conf --verbose
read template file test/bitcoin-template.conf
written output file test/bitcoin-output.conf
3 replaced, 0 kept, 0 deleted, 2 lines deleted, 2 set to default

$ cat test/bitcoin-output.conf
# network
mainnet=1
testnet=0
rpcconnect=127.0.0.1
dbcache=300
seednode=nkf5e6b7pl4jfd4a.onion
```

## Testing

The following files are used for automated testing (not yet implemented), TODO(Stadicus):

* `test-redisimport.txt`: bulk import key/value pairs into Redis
* `test-template.conf`: template config file
* `test-reference.conf`: reference config file, to compare newly created `test-output.conf` with
