#!/usr/bin/env bash

ANNPATH="/home/kyli/Codes/golang/src/github.com/dappledger/AnnChain/build"
CHAINID="annchain-gGgTjd"

${ANNPATH}/anntool --backend="tcp://127.0.0.1:16657" --target="${CHAINID}" event request --privkey="${PRIV1}" --source="evm2" --listener="ikhofi1" --source_hash="d4951d2b9cb115f8653f10200d26edc6c35765e2ba660a041a2971bff0f8509a"  --listener_hash="03dbe24c98d6da7d66446ef2af5a53dd4fb6b4f09688e5a7590b59ba2e49b863"
