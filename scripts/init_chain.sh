#!/usr/bin/env bash

t=$1
n=$2

rm -rf $HOME/.dkgcli

dir_path=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )
herb_path=$dir_path/..
bots_path=$herb_path/bots

pwrd="alicealice"

dkgcli gen-key-file $t $n

cd $HOME/.dkgcli

ck=$(cat keys.json | jq .common_key)
Ck=${ck:1:${#ck}-2}

rm -rf $HOME/.hcli
rm -rf $HOME/.hd

rm -rf $bots_path
mkdir $bots_path

hd init moniker --chain-id HERBchain

for (( i=0; i<$n; i++ ))
do
    cd $HOME/.dkgcli

    hcli keys add "node$i" <<< $pwrd

    hd add-genesis-account $(hcli keys show "node$i" -a) 1000herbtoken,100000000stake

    id=$(cat keys.json | jq .partial_keys[$i].id)
    ID=${id:1:${#id}-2}

    vk=$(cat keys.json | jq .partial_keys[$i].verification_key)
    Vk=${vk:1:${#vk}-2}

    hd add-key-holder $(hcli keys show "node$i" -a) $ID $Vk

    pk=$(cat keys.json | jq .partial_keys[$i].private_key)
    Pk=${pk:1:${#pk}-2}

    cd $bots_path

    echo "#!/usr/bin/expect -f
      set timeout -1
      cd $dir_path

      spawn ./HERB node$i $Pk $ID $Ck

      match_max 100000

      while { true } {
        expect \"Password to sign with 'node$i':\"
        send -- \"alicealice\r\"
      }

      expect eof" > "node$i".exp

    chmod +x ./"node$i".exp

done

hd set-threshold $t $t
hd set-common-key $Ck

hcli config chain-id HERBchain
hcli config output json
hcli config indent true
hcli config trust-node true

hd gentx --name node0 <<< $pwrd

hd collect-gentxs

hd validate-genesis