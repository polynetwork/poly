module github.com/polynetwork/poly

go 1.14

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/cosmos/cosmos-sdk v0.39.0
	github.com/ethereum/go-ethereum v1.9.15
	github.com/gcash/bchd v0.16.5
	github.com/gcash/bchutil v0.0.0-20200506001747-c2894cd54b33
	github.com/gorilla/websocket v1.4.2
	github.com/gosuri/uiprogress v0.0.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/itchyny/base58-go v0.1.0
	github.com/joeqian10/neo-gogogo v0.0.0-20200716075409-923bd4879b43
	github.com/ontio/ontology v1.11.0
	github.com/ontio/ontology-crypto v1.0.9
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/pborman/uuid v1.2.0
	github.com/stretchr/testify v1.6.1
	github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d
	github.com/tendermint/tendermint v0.33.6
	github.com/tjfoc/gmsm v1.3.2-0.20200914155643-24d14c7bd05c
	github.com/urfave/cli v1.22.4
	github.com/valyala/bytebufferpool v1.0.0
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	gotest.tools v2.2.0+incompatible
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta3.0.20201006151309-9c426dcc5096
	github.com/hyperledger/fabric v1.4.3
)

replace github.com/tjfoc/gmsm => github.com/zouxyan/gmsm v1.3.2-0.20200925082225-a66aabdb8da8
