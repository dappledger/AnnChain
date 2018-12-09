#!/usr/bin/env bash

ANNPATH="/home/kyli/Codes/golang/src/github.com/dappledger/AnnChain/build"
CHAINID="annchain-gGgTjd"

PRIV1="9F04A3EB2E3B412617F0A9D39466B357EBD3A073C28D004C73E482544515898D0FC4E216FB4B40781CEFAECB6C359BA6549069475B7DD678AECF1DF4AC5FCB4E"
PRIV2="2E28B2B44378E96048C777B1D48E5FB2AA3D88F295CC6D3371938C5B7820C8C47B039128642136EF7303A5CFAB5C719F705FCF1C08CA981DDEDCE1001932E12E"
PRIV3="CC1584C9ED0B56A5E1232E5F29425FE9E164EE01C5CA289F6017ED7143374867BC8DDD6A76E328F557069DBEA251D752F5EB284BF97583D2E37FD600783E83ED"
PRIV4="CBAF971D0B9CB4DD005A35878D5A31A5D9C27181EAE1091AC3300A80D6108076C8722AB9C75E00817463AF0BD2AD95B10A95BC3493B494B7016E1EC224D812D2"

${ANNPATH}/anntool --callmode="commit" --backend="tcp://127.0.0.1:16657" --target="${CHAINID}" event upload-code --privkey="${PRIV1}" --code 'function main(params)
if params["contract_call"] == nil or params.contract_call["function"] ~= "buyChicken" then
    return nil;
end

return {
    ["from"] = params.from,
    ["to"] = params.to,
    ["value"] = params.value,
    ["nonce"] = params.nonce,
    ["function"] = params.contract_call["function"],
    ["score"] = params.contract_call["_score"]
};
end
' --ownerid evm2

${ANNPATH}/anntool --backend="tcp://127.0.0.1:16657" --callmode="commit" --target="${CHAINID}" event upload-code --privkey="${PRIV1}" --code 'function main(params)
if tonumber(params.score) < 1000 then
   return nil;
else
   return params;
end
end
' --ownerid ikhofi1
