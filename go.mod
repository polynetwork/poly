module github.com/polynetwork/poly

go 1.14

require (
	github.com/Zilliqa/gozilliqa-sdk v1.2.1-0.20210927032600-4c733f2cb879
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/confio/ics23/go v0.6.6
	github.com/cosmos/cosmos-sdk v0.39.1
	github.com/ethereum/go-ethereum v1.9.25
	github.com/gcash/bchd v0.16.5
	github.com/gcash/bchutil v0.0.0-20200506001747-c2894cd54b33
	github.com/gorilla/websocket v1.4.2
	github.com/gosuri/uiprogress v0.0.1
	github.com/harmony-one/bls v0.0.6
	github.com/harmony-one/harmony v1.10.3-0.20220216090956-7e6b16aec8dc
	github.com/hashicorp/golang-lru v0.5.4
	github.com/holiman/uint256 v1.2.0
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/itchyny/base58-go v0.1.0
	github.com/joeqian10/neo-gogogo v1.1.0
	github.com/joeqian10/neo3-gogogo v0.3.8
	github.com/joeqian10/neo3-gogogo-legacy v1.0.0
	github.com/novifinancial/serde-reflection/serde-generate/runtime/golang v0.0.0-20210526181959-1694c58d103e
	github.com/ontio/ontology v1.11.1-0.20200812075204-26cf1fa5dd47
	github.com/ontio/ontology-crypto v1.0.9
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/polynetwork/poly-io-test v0.0.0-20200819093740-8cf514b07750
	github.com/polynetwork/ripple-sdk v1.0.0
	github.com/renlulu/gozilliqa-sdklegacy v0.0.0-20220127085552-852a2675dc93
	github.com/rubblelabs/ripple v0.0.0-20220222071018-38c1a8b14c18
	github.com/starcoinorg/starcoin-go v0.0.0-20220105024102-530daedc128b
	github.com/stretchr/testify v1.7.0
	github.com/switcheo/tendermint v0.34.14-2
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.7
	github.com/urfave/cli v1.22.4
	github.com/valyala/bytebufferpool v1.0.0
	golang.org/x/crypto v0.0.0-20211117183948-ae814b36b871
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	gotest.tools v2.2.0+incompatible
)

replace (
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce => github.com/btcsuite/btcutil v1.0.2
	github.com/ethereum/go-ethereum v1.9.25 => github.com/ethereum/go-ethereum v1.9.15
	github.com/harmony-one/harmony v1.10.3-0.20220216090956-7e6b16aec8dc => github.com/devfans/harmony v1.10.3-0.20220304055439-856e256b615f
	github.com/polynetwork/ripple-sdk v1.0.0 => github.com/siovanus/ripple-sdk v1.0.1-0.20220311082414-84e86a29df1a
	github.com/rubblelabs/ripple v0.0.0-20220222071018-38c1a8b14c18 => github.com/siovanus/ripple v0.0.0-20220311080636-cbff6a9e07ce
	github.com/tendermint/tm-db/064 => github.com/tendermint/tm-db v0.6.4
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf => golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
)
