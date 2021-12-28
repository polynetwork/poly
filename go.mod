module github.com/polynetwork/poly

go 1.14

require (
	github.com/Zilliqa/gozilliqa-sdk v1.2.1-0.20210927032600-4c733f2cb879
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/confio/ics23/go v0.6.6
	github.com/cosmos/cosmos-sdk v0.39.1
	github.com/ethereum/go-ethereum v1.9.15
	github.com/gcash/bchd v0.16.5
	github.com/gcash/bchutil v0.0.0-20200506001747-c2894cd54b33
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/gosuri/uiprogress v0.0.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/itchyny/base58-go v0.1.0
	github.com/joeqian10/neo-gogogo v1.1.0
	github.com/joeqian10/neo3-gogogo v0.3.8
	github.com/joeqian10/neo3-gogogo-legacy v1.0.0
	github.com/kr/pretty v0.2.1 // indirect
	github.com/ontio/ontology v1.11.1-0.20200812075204-26cf1fa5dd47
	github.com/ontio/ontology-crypto v1.0.9
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/polynetwork/poly-io-test v0.0.0-20200819093740-8cf514b07750
	github.com/prometheus/client_golang v1.8.0 // indirect
	github.com/renlulu/gozilliqa-sdklegacy v0.0.0-20210926114807-88a08c5ab803
	github.com/spf13/cobra v1.1.1 // indirect
	github.com/spf13/viper v1.7.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.7
	github.com/urfave/cli v1.22.4
	github.com/valyala/bytebufferpool v1.0.0
	go.etcd.io/bbolt v1.3.5 // indirect
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/net v0.0.0-20210903162142-ad29c8ab022f
	golang.org/x/sys v0.0.0-20210903071746-97244b99971b // indirect
	google.golang.org/grpc v1.37.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gotest.tools v2.2.0+incompatible
)

replace github.com/tendermint/tm-db/064 => github.com/tendermint/tm-db v0.6.4

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
