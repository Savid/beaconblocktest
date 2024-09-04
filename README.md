# beaconblocktest

Check the size of beacon blocks. The sizes are calculated from the `data.message` key in the json response from the [getBlockV2](https://ethereum.github.io/beacon-APIs/#/Beacon/getBlockV2) beacon API.

- `raw` - use raw HTTP request to the beacon node
- `eth2client` - use [`eth2-client`](https://github.com/attestantio/go-eth2-client) library

## Usage

`go run main.go <slot_number>`

Example:
```
# default localhost:5052
go run main.go 9882669

# beacon node on 192.168.1.1:5050
BEACON_ADDR=192.168.1.1 PORT=5050 go run main.go 9882669
```
