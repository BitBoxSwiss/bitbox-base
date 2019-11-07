# bbbconfgen

Application to generate text files from a template, replacing placeholders that specify Redis keys.
It's written in Go as part of the [BitBoxBase](https://github.com/digitalbitbox/bitbox-base) project by [Shift Cryptosecurity](https://shiftcrypto.ch) and used for automatically generating configuration files.

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
  --version
  --quiet
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

The #rmLineTrue and #rmLineFalse functions allows to drop a line conditionally.

  {{ key #rmLineTrue }}         drop line if a key is set to '1', 'true', 'yes' or 'y'
  {{ key #rmLineFalse }}        drop line if a key is set to '0', 'false', 'no', 'n' or not at all
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
$ redis-cli SET tor:enabled 1
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
printtoconsole=1                                        {{ bitcoind:mainnet #rmLineTrue }}
seednode={{ bitcoind:seednode:1 #rmLine }}              {{ tor:enabled #rmLineFalse }}
seednode={{ bitcoind:seednode:2 #rmLine }}              {{ tor:enabled #rmLineFalse }}
seednode={{ bitcoind:seednode:3 #rmLine }}              {{ tor:enabled #rmLineFalse }}

$ ./bbbconfgen --template test/bitcoin-template.conf --output test/bitcoin-output.conf
connected to Redis
opened template config file test/bitcoin-template.conf
writing into output file test/bitcoin-output.conf
written 8 lines
placeholders: 7 replaced, 0 kept, 0 deleted, 0 lines deleted, 0 set to default
checks: 1 lines dropped, 3 lines kept

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
