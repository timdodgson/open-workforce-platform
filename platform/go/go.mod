module github.com/timdodgson/open-workforce-platform/platform/go

go 1.26.4

require (
	github.com/aws/aws-sdk-go-v2/config v1.32.36
	github.com/aws/aws-sdk-go-v2/service/s3 v1.104.2
	github.com/timdodgson/open-workforce-platform/examples/byod-tsp v0.0.0
	github.com/timdodgson/open-workforce-platform/owp-sdk v0.1.0
)

replace (
	github.com/timdodgson/open-workforce-platform/examples/byod-tsp => ../../examples/byod-tsp
	github.com/timdodgson/open-workforce-platform/owp-sdk => ../owp-sdk
)

require (
	github.com/aws/aws-sdk-go-v2 v1.43.5 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.14 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.35 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.36 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.31 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.5.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.33.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.38.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.45.5 // indirect
	github.com/aws/smithy-go v1.27.7 // indirect
)
