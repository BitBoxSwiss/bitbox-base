#!/usr/bin/env python3
# Copyright (c) 2015-2018 The Bitcoin Core developers
# Distributed under the MIT software license, see the accompanying
# file COPYING or http://www.opensource.org/licenses/mit-license.php.
# 
# Original script
# https://github.com/bitcoin/bitcoin/tree/master/share/rpcauth
# ------------------------------------------------------------------------------
# Extendet script, by Shift Cryptosecurity AG, Switzerland
# implementing direct Redis support to store rpcauth, rpcuser and rpcpassword
# 
# https://github.com/digitalbitbox/bitbox-base
# ------------------------------------------------------------------------------

from argparse import ArgumentParser
from base64 import urlsafe_b64encode
from binascii import hexlify
from getpass import getpass
from os import urandom
import redis
import sys

import hmac

def generate_salt(size):
    """Create size byte hex salt"""
    return hexlify(urandom(size)).decode()

def generate_password():
    """Create 32 byte b64 password"""
    return urlsafe_b64encode(urandom(32)).decode('utf-8')

def password_to_hmac(salt, password):
    m = hmac.new(bytearray(salt, 'utf-8'), bytearray(password, 'utf-8'), 'SHA256')
    return m.hexdigest()

def main():
    parser = ArgumentParser(description='Create login credentials for a JSON-RPC user')
    parser.add_argument('username', help='the username for authentication')
    parser.add_argument('password', help='leave empty to generate a random password or specify "-" to prompt for password', nargs='?')
    args = parser.parse_args()

    if not args.password:
        args.password = generate_password()
    elif args.password == '-':
        args.password = getpass()

    # Create 16 byte hex salt
    salt = generate_salt(16)
    password_hmac = password_to_hmac(salt, args.password)

    print('String to be appended to bitcoin.conf:')
    print('rpcauth={0}:{1}${2}'.format(args.username, salt, password_hmac))
    print('Your password:\n{0}'.format(args.password))

    # Extension by Shift Cryptosecurity AG, Switzerland 
    # ▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼

    try:
        r = redis.Redis(
            host='127.0.0.1',
            port=6379,
        )

        if r.ping():
            r.set('bitcoind:rpcauth', '{0}:{1}${2}'.format(args.username, salt, password_hmac))
            r.set('bitcoind:rpcuser', args.username)
            r.set('bitcoind:rpcpassword', args.password)
            r.save()
        else:
            sys.exit('ERR: Redis not available.')

    except:
        sys.exit('ERR: Redis not available.')

    # ▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲

if __name__ == '__main__':
    main()
