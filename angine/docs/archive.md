# Data Archive

Archive block data can reduce storage space, and facilitate new node join the cluster

archive object: block data <br/>
ticapsule: distributed file storage service, make sure storaged data not to be tampered

## Usage

Add 4 items in your config file such as `~/.angine/config.toml`:

ti_endpoint = `<string|ticapsule url>` <br/>
ti_key = `<string|ticapsule api key>` <br/>
ti_secret = `<string|ticapsule secret key>`<br/>
threshold_blocks = `<int|threshold for triggering upload>`

