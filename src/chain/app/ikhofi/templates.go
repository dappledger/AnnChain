package ikhofi

const CONFIGTPL = `# This is a YAML config file.
---
# default config for ikhofi
# database
db:
type: leveldb
path: /.ann_runtime/data/chaindata/
cacheSize: 67108864 # 64MB
destroyAfterClose: false
`
