module github.com/alibaba/higress/plugins/wasm-go/extensions/request-validation

go 1.19

replace github.com/alibaba/higress/plugins/wasm-go => ../..

require (
	github.com/alibaba/higress/plugins/wasm-go v1.3.1
	github.com/higress-group/proxy-wasm-go-sdk v0.0.0-20240318034951-d5306e367c43
	github.com/santhosh-tekuri/jsonschema v1.2.4
	github.com/tidwall/gjson v1.17.0
)

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/higress-group/nottinygc v0.0.0-20231101025119-e93c4c2f8520 // indirect
	github.com/magefile/mage v1.14.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/resp v0.1.1 // indirect
)