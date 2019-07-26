cd C:/
echo y | rmdir /s .hcli
echo y | rmdir /s .hd
cd go-path/bin
hd init ilia --chain-id HERBchain
(echo alicealice && echo alicealice) | hcli keys add alice
(echo alicealice && echo alicealice) | hcli keys add bob
(echo alicealice && echo alicealice) | hcli keys add carol
for /f "delims=" %%a in ('hcli keys show alice -a') do @set aliceAddr=%%a
for /f "delims=" %%a in ('hcli keys show bob -a') do @set bobAddr=%%a
for /f "delims=" %%a in ('hcli keys show carol -a') do @set carolAddr=%%a
hd add-genesis-account %aliceAddr% 1000herbtoken,100000000stake
hd add-genesis-account %bobAddr% 1000herbtoken,100000000stake
hd add-genesis-account %carolAddr% 1000herbtoken,100000000stake
hd set-threshold 3 2
hd set-common-key 0452b4a6d7883102258a87539c41898cd1c78bcc27dd905d9111e8b066504ba31b160580530886a2200833c2281e10377dbb2007abc531959a23df365ffc16ee18
hd add-key-holder %aliceAddr% 0 041ea050368a68a13a12f1026870b997d6d15d74a59f243ef9c38aea5089387f13dd754344b73ab5d59a716a13abcf4cc3767b723a60d1c367dde5b52d3b04781f
hd add-key-holder %bobAddr% 1 04711451d30356c470e156941119d44cd4e8b44f04497c1875be584c93d423405e96f8fe4efb8c7d181bdea18b2ba1673f0d639eba187e491ae00d25320711f9f5
hd add-key-holder %carolAddr% 2 04e90f797b084978e896f398dd408249d72218803f155044c994c4f4ac7da57db44e00771ecec98e3a213a9e141e678700011d923ab768d5d329916c611dea85cf
hcli config chain-id HERBchain
hcli config output json
hcli config indent true
hcli config trust-node true
echo alicealice | hd gentx --name alice
hd collect-gentxs
hd validate-genesis
PAUSE