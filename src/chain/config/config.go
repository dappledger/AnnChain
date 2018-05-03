package config

var Template = `#toml configuration for annchain
environment = "development"               # log mode, e.g. "development"/"production"
rpc_laddr = "0.0.0.0:46657"               # rpc port this node is exposing
event_laddr = "0.0.0.0:46658"             # annchain uses a exposed port for events function
log_path = ""                             # 
`
