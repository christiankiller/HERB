#!/bin/bash
rm -rf ~/.hcli
rm -rf ~/.hd
rm -rf ~/HERB/bots
mkdir ~/HERB/bots
hd init moniker --chain-id HERBchain
hcli config chain-id HERBchain
hcli config output json
hcli config indent true
hcli config trust-node true
cd $HOME
mkdir -p $HOME/.hd/config
cp genesis.json .hd/config
cp config.toml .hd/config
cp -r keys .hcli/
cp -r bots HERB/
sed -i 's/moniker = "moniker"/moniker = "node-'"$1"'"/' .hd/config/config.toml
