anntool sign -sec 36367998C2AB07EFBB92AE9299108E0C6375597C925AB17515AD00A78CB20C578FC3DB8629DC8ABCC14B0967F354329D6151D0EBA4B5DDD3E0644FA86A846775 -pub 8FC3DB8629DC8ABCC14B0967F354329D6151D0EBA4B5DDD3E0644FA86A846775test-chain-2,548CADCBF01F531F9FC4C690A8297632E1DE1E902D3CFF68747317B03FE49EB8test-chain-2


EVM:
/home/kyli/Codes/golang/src/github.com/dappledger/AnnChain/build/anntool  --backend=tcp://127.0.0.1:16657 --target annchain-qWJdh9 organization create --genesisfile ./shards/test-chain-2/genesis.json --configfile ./shards/test-chain-2/config1.json --privkey 36367998C2AB07EFBB92AE9299108E0C6375597C925AB17515AD00A78CB20C578FC3DB8629DC8ABCC14B0967F354329D6151D0EBA4B5DDD3E0644FA86A846775 --pubkey 8FC3DB8629DC8ABCC14B0967F354329D6151D0EBA4B5DDD3E0644FA86A846775

/home/kyli/Codes/golang/src/github.com/dappledger/AnnChain/build/anntool  --backend=tcp://127.0.0.1:26657 --target annchain-qWJdh9 organization join --genesisfile ./shards/test-chain-2/genesis.json --configfile ./shards/test-chain-2/config2.json --privkey C1AD689716CD6C592E69ADD556EA7E1517A3D8C3FBCBA84ED2E6FE6372C944D0548CADCBF01F531F9FC4C690A8297632E1DE1E902D3CFF68747317B03FE49EB8 --pubkey 548CADCBF01F531F9FC4C690A8297632E1DE1E902D3CFF68747317B03FE49EB8

/home/kyli/Codes/golang/src/github.com/dappledger/AnnChain/build/anntool  --backend=tcp://127.0.0.1:36657 --target annchain-qWJdh9 organization join --genesisfile ./shards/test-chain-2/genesis.json --configfile ./shards/test-chain-2/config3.json --privkey 79209FFA9E31045B864B92B707FA55894FF53BE486A9A21B22205A012C76D1040F88C125D90B2A980457039609D9D2599352C0C5115C88209DC64911CA2213CF --pubkey 0F88C125D90B2A980457039609D9D2599352C0C5115C88209DC64911CA2213CF

/home/kyli/Codes/golang/src/github.com/dappledger/AnnChain/build/anntool --backend=tcp://127.0.0.1:46657 --target annchain-qWJdh9 organization join --genesisfile ./shards/test-chain-2/genesis.json --configfile ./shards/test-chain-2/config4.json --privkey BC58D786781D733666F6B1979CC1794DB711666B0CAFD58417E31C00FE1D1206F4377A4B617207415D9D1A5349908DB1BC95D0D2BA1BFF08371E1EAA3194ABB9 --pubkey F4377A4B617207415D9D1A5349908DB1BC95D0D2BA1BFF08371E1EAA3194ABB9


SpecialOP:
/home/kyli/Codes/golang/src/github.com/dappledger/AnnChain/build/anntool --backend=tcp://127.0.0.1:16657 --chainid test-chain-2 special change_validator --account_pubkey aasdfasdfasdfas --validator_pubkey 8FC3DB8629DC8ABCC14B0967F354329D6151D0EBA4B5DDD3E0644FA86A846775 --power 100 --isCA true --rpc fuckyou

Event request:
/home/kyli/Codes/golang/src/github.com/dappledger/AnnChain/build/anntool --backend="tcp://127.0.0.1:16657" --target annchain-qWJdh9 event request --privkey 36367998C2AB07EFBB92AE9299108E0C6375597C925AB17515AD00A78CB20C578FC3DB8629DC8ABCC14B0967F354329D6151D0EBA4B5DDD3E0644FA86A846775 --pubkey 8FC3DB8629DC8ABCC14B0967F354329D6151D0EBA4B5DDD3E0644FA86A846775 --listener "test-chain-2" --source "annchain-qWJdh9"


anntool --backend "tcp://127.0.0.1:16657" query nonce --chainid "test-chain-2" --address 0x7752b42608a0f1943c19fc5802cb027e60b4c911

anntool --backend "tcp://127.0.0.1:16657" contract create --chainid "test-chain-2" --abif ./sample.abi --callf ./sample.deploy.json --nonce 1

anntool --backend "tcp://127.0.0.1:16657" contract read --chainid "test-chain-2" --callf ./sample.getowner.json --abif ./sample.abi --nonce 2
