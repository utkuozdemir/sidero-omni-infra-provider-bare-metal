module github.com/siderolabs/omni-infra-provider-bare-metal

go 1.23.2

replace (
	github.com/pensando/goipmi v0.0.0-20240603174436-eb122d901c23 => github.com/siderolabs/goipmi v0.0.0-20211214143420-35f956689e67
	github.com/pin/tftp/v3 v3.1.0 => github.com/utkuozdemir/pin-tftp/v3 v3.0.0-20241021135417-0dd7dba351ad
	github.com/siderolabs/omni/client v0.0.0-20241017162757-284e8b5077cc => github.com/utkuozdemir/sidero-omni/client v0.0.0-20241018221630-bf35a3add198
	github.com/siderolabs/talos-metal-agent v0.0.0-20241016074728-46df49991336 => github.com/utkuozdemir/sidero-talos-metal-agent v0.0.0-20241023212532-42f37aaacfad
)

require (
	github.com/cosi-project/runtime v0.6.4
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0
	github.com/insomniacslk/dhcp v0.0.0-20240829085014-a3a4c1f04475
	github.com/jhump/grpctunnel v0.3.0
	github.com/pensando/goipmi v0.0.0-20240603174436-eb122d901c23
	github.com/pin/tftp/v3 v3.1.0
	github.com/planetscale/vtprotobuf v0.6.1-0.20240917153116-6f2963f01587
	github.com/siderolabs/gen v0.5.0
	github.com/siderolabs/omni/client v0.0.0-20241017162757-284e8b5077cc
	github.com/siderolabs/talos-metal-agent v0.0.0-20241016074728-46df49991336
	github.com/spf13/cobra v1.8.1
	go.uber.org/zap v1.27.0
	golang.org/x/net v0.30.0
	golang.org/x/sync v0.8.0
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
)

require (
	github.com/ProtonMail/go-crypto v1.1.0-alpha.5.0.20240827111422-b5837fa4476e // indirect
	github.com/ProtonMail/go-mime v0.0.0-20230322103455-7d82a3887f2f // indirect
	github.com/ProtonMail/gopenpgp/v2 v2.7.5 // indirect
	github.com/adrg/xdg v0.5.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cloudflare/circl v1.3.9 // indirect
	github.com/containerd/go-cni v1.1.10 // indirect
	github.com/containernetworking/cni v1.2.3 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fullstorydev/grpchan v1.1.1 // indirect
	github.com/gertd/go-pluralize v0.2.1 // indirect
	github.com/google/cel-go v0.21.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/jsimonetti/rtnetlink/v2 v2.0.2 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mdlayher/ethtool v0.1.0 // indirect
	github.com/mdlayher/genetlink v1.3.2 // indirect
	github.com/mdlayher/netlink v1.7.2 // indirect
	github.com/mdlayher/socket v0.5.1 // indirect
	github.com/opencontainers/runtime-spec v1.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/siderolabs/crypto v0.4.4 // indirect
	github.com/siderolabs/go-api-signature v0.3.6 // indirect
	github.com/siderolabs/go-blockdevice v0.4.7 // indirect
	github.com/siderolabs/go-blockdevice/v2 v2.0.3 // indirect
	github.com/siderolabs/go-pointer v1.0.0 // indirect
	github.com/siderolabs/image-factory v0.5.0 // indirect
	github.com/siderolabs/net v0.4.0 // indirect
	github.com/siderolabs/proto-codec v0.1.1 // indirect
	github.com/siderolabs/protoenc v0.2.1 // indirect
	github.com/siderolabs/talos/pkg/machinery v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/u-root/uio v0.0.0-20240209044354-b3d14b93376a // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241015192408-796eee8c2d53 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
