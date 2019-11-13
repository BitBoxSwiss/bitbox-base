![BitBoxBase logo](bitbox-base-logo.png)

[![Build Status](https://travis-ci.org/digitalbitbox/bitbox-base.svg?branch=master)](https://travis-ci.org/digitalbitbox/bitbox-base)

## Personal Bitcoin & Lightning node

The BitBoxBase is an ongoing project of [Shift Cryptosecurity](https://shiftcrypto.ch/) that aims to build a personal Bitcoin full node appliance.
The whole software stack is free open-source.
This documentation is aimed at project members, contributors and intersted people that want to build or customize their own node.

## Documentation

Detailed documentation is available at <https://base.shiftcrypto.ch>.

## Why run a Bitcoin node

We believe that storing Bitcoin private keys on a hardware wallet like our [BitBox](https://shiftcrypto.ch) is only one part of the equation to gain financial sovereignty.
While hardware wallets provide security, they do not provide privacy.
Your entire financial history can be read by the company, such as the hardware wallet provider, who querries the blockchain for you.

The currently missing part of the equation is a personal appliance that syncs directly with the Bitcoin peer-to-peer network and is able to send and validate transactions in a private manner.
Because we respect an individual's right to privacy, we decided to build the BitBoxBase.

Running a Bitcoin node makes you a direct network participant, giving you additional security and privacy.
And Bitcoin as a decentralized system is better off with it (see [blog post](https://medium.com/shiftcrypto/we-need-bitcoin-full-nodes-economic-ones-fd17efcb61fb) for additional details).

## Our goals

Running your own Bitcoin node in combination with a hardware wallet is still to complicated.
By building the BitBoxBase, we want to achieve the following goals:

* Running your own Bitcoin full node is for everyone.
* The built-in Lightning client provides a compelling Lightning Wallet in the BitBoxApp.
* Connecting to your node just works, whether in your own network or on-the-go.
* Privacy is assured through end-to-end encryption between User Interface and BitBoxBase.
* As a networked appliance, remote attack surface is minimized by exposing as little ports as possible.
* The hardware platform uses best-in-class components, built for performance and resilience.
* With the integrated BitBox secure module, the node offers functionality previously not possible with hardware wallets.
* Atomic upgrades allows seamless and reliable Base image upgrades with fallback.
* Expert settings allow access to low-level configuration.

## Buy or Build

We strive to build a professional Bitcoin node as part of our our Shift Cryptosecurity product portfolio, working seamlessly with the BitBox hardware wallet and BitBoxApp.
Users will be able to buy it, and receive professional support and maintenance.

The overarching goal, however, is to enable everyone to run a Bitcoin full node. This is why you can also build it yourself, with standard parts and completely open-source code.

## Contributor workflow

We are building the software stack of the BitBoxBase fully open source and with its application outside of our own hardware device in mind.
Contributions are very welcome.
Please read [CONTRIBUTING](CONTRIBUTING.md) before submitting changes to the repository.
