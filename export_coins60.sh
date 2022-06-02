#!/bin/sh
curl -s https://r.easycrypto.nz/json/public/coins60.csv | jq  ".coins[] | { name: .name, symbol: .symbol}" | jq -r '"\"" + (.symbol) + "\": \""  +  (.name) + "\""' > assets.yaml
