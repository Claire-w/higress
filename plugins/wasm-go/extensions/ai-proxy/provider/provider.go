package provider

import (
	"errors"
	"math/rand"

	"github.com/alibaba/higress/plugins/wasm-go/pkg/wrapper"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/tidwall/gjson"
)

type ApiName string
type Pointcut string

const (
	ApiNameChatCompletion ApiName = "chatCompletion"

	providerTypeMoonshot = "moonshot"
	providerTypeAzure    = "azure"
	providerTypeQwen     = "qwen"
	providerTypeOpenAI   = "openai"
	providerTypeGroq     = "groq"
	providerTypeBaichuan = "baichuan"
	providerTypeYi       = "yi"
	providerTypeDeepSeek = "deepseek"
	providerTypeOllama   = "ollama"

	protocolOpenAI   = "openai"
	protocolOriginal = "original"

	roleSystem = "system"

	ctxKeyIncrementalStreaming = "incrementalStreaming"
	ctxKeyStreamingBody        = "streamingBody"
	ctxKeyOriginalRequestModel = "originalRequestModel"
	ctxKeyFinalRequestModel    = "finalRequestModel"
	ctxKeyPushedMessage        = "pushedMessage"

	objectChatCompletion      = "chat.completion"
	objectChatCompletionChunk = "chat.completion.chunk"

	wildcard = "*"

	defaultTimeout = 2 * 60 * 1000 // ms
)

type providerInitializer interface {
	ValidateConfig(ProviderConfig) error
	CreateProvider(ProviderConfig) (Provider, error)
}

var (
	errUnsupportedApiName = errors.New("unsupported API name")

	providerInitializers = map[string]providerInitializer{
		providerTypeMoonshot: &moonshotProviderInitializer{},
		providerTypeAzure:    &azureProviderInitializer{},
		providerTypeQwen:     &qwenProviderInitializer{},
		providerTypeOpenAI:   &openaiProviderInitializer{},
		providerTypeGroq:     &groqProviderInitializer{},
		providerTypeBaichuan: &baichuanProviderInitializer{},
		providerTypeYi:       &yiProviderInitializer{},
		providerTypeDeepSeek: &deepseekProviderInitializer{},
		providerTypeOllama:   &ollamaProviderInitializer{},
	}
)

type Provider interface {
	GetProviderType() string
}

type RequestHeadersHandler interface {
	OnRequestHeaders(ctx wrapper.HttpContext, apiName ApiName, log wrapper.Log) (types.Action, error)
}

type RequestBodyHandler interface {
	OnRequestBody(ctx wrapper.HttpContext, apiName ApiName, body []byte, log wrapper.Log) (types.Action, error)
}

type ResponseHeadersHandler interface {
	OnResponseHeaders(ctx wrapper.HttpContext, apiName ApiName, log wrapper.Log) (types.Action, error)
}

type StreamingResponseBodyHandler interface {
	OnStreamingResponseBody(ctx wrapper.HttpContext, name ApiName, chunk []byte, isLastChunk bool, log wrapper.Log) ([]byte, error)
}

type ResponseBodyHandler interface {
	OnResponseBody(ctx wrapper.HttpContext, apiName ApiName, body []byte, log wrapper.Log) (types.Action, error)
}

type ProviderConfig struct {
	// @Title zh-CN AI服务提供商
	// @Description zh-CN AI服务提供商类型，目前支持的取值为："moonshot"、"qwen"、"openai"、"azure"、"baichuan"、"yi"、"ollama"
	typ string `required:"true" yaml:"type" json:"type"`
	// @Title zh-CN API Tokens
	// @Description zh-CN 在请求AI服务时用于认证的API Token列表。不同的AI服务提供商可能有不同的名称。部分供应商只支持配置一个API Token（如Azure OpenAI）。
	apiTokens []string `required:"false" yaml:"apiToken" json:"apiTokens"`
	// @Title zh-CN 请求超时
	// @Description zh-CN 请求AI服务的超时时间，单位为毫秒。默认值为120000，即2分钟
	timeout uint32 `required:"false" yaml:"timeout" json:"timeout"`
	// @Title zh-CN Moonshot File ID
	// @Description zh-CN 仅适用于Moonshot AI服务。Moonshot AI服务的文件ID，其内容用于补充AI请求上下文
	moonshotFileId string `required:"false" yaml:"moonshotFileId" json:"moonshotFileId"`
	// @Title zh-CN Azure OpenAI Service URL
	// @Description zh-CN 仅适用于Azure OpenAI服务。要请求的OpenAI服务的完整URL，包含api-version等参数
	azureServiceUrl string `required:"false" yaml:"azureServiceUrl" json:"azureServiceUrl"`
	// @Title zh-CN 通义千问File ID
	// @Description zh-CN 仅适用于通义千问服务。上传到Dashscope的文件ID，其内容用于补充AI请求上下文。仅支持qwen-long模型。
	qwenFileIds []string `required:"false" yaml:"qwenFileIds" json:"qwenFileIds"`
	// @Title zh-CN 启用通义千问搜索服务
	// @Description zh-CN 仅适用于通义千问服务，表示是否启用通义千问的互联网搜索功能。
	qwenEnableSearch bool `required:"false" yaml:"qwenEnableSearch" json:"qwenEnableSearch"`
	// @Title zh-CN Ollama Server IP/Domain
	// @Description zh-CN 仅适用于 Ollama 服务。Ollama 服务器的 IP 地址。
	ollamaServerIP string `required:"false" yaml:"ollamaServerIP" json:"ollamaServerIP"`
	// @Title zh-CN Ollama Server Port
	// @Description zh-CN 仅适用于 Ollama 服务。Ollama 服务器的端口号。
	ollamaServerPort uint32 `required:"false" yaml:"ollamaServerPort" json:"ollamaServerPort"`
	// @Title zh-CN 模型名称映射表
	// @Description zh-CN 用于将请求中的模型名称映射为目标AI服务商支持的模型名称。支持通过“*”来配置全局映射
	modelMapping map[string]string `required:"false" yaml:"modelMapping" json:"modelMapping"`
	// @Title zh-CN 对外接口协议
	// @Description zh-CN 通过本插件对外提供的AI服务接口协议。默认值为“openai”，即OpenAI的接口协议。如需保留原有接口协议，可配置为“original"
	protocol string `required:"false" yaml:"protocol" json:"protocol"`
	// @Title zh-CN 模型对话上下文
	// @Description zh-CN 配置一个外部获取对话上下文的文件来源，用于在AI请求中补充对话上下文
	context *ContextConfig `required:"false" yaml:"context" json:"context"`
}

func (c *ProviderConfig) FromJson(json gjson.Result) {
	c.typ = json.Get("type").String()
	c.apiTokens = make([]string, 0)
	for _, token := range json.Get("apiTokens").Array() {
		c.apiTokens = append(c.apiTokens, token.String())
	}
	c.timeout = uint32(json.Get("timeout").Uint())
	if c.timeout == 0 {
		c.timeout = defaultTimeout
	}
	c.moonshotFileId = json.Get("moonshotFileId").String()
	c.azureServiceUrl = json.Get("azureServiceUrl").String()
	c.qwenFileIds = make([]string, 0)
	for _, fileId := range json.Get("qwenFileIds").Array() {
		c.qwenFileIds = append(c.qwenFileIds, fileId.String())
	}
	c.qwenEnableSearch = json.Get("qwenEnableSearch").Bool()
	c.ollamaServerIP = json.Get("ollamaServerIP").String()
	c.ollamaServerPort = uint32(json.Get("ollamaServerPort").Uint())
	c.modelMapping = make(map[string]string)
	for k, v := range json.Get("modelMapping").Map() {
		c.modelMapping[k] = v.String()
	}
	c.protocol = json.Get("protocol").String()
	if c.protocol == "" {
		c.protocol = protocolOpenAI
	}
	contextJson := json.Get("context")
	if contextJson.Exists() {
		c.context = &ContextConfig{}
		c.context.FromJson(contextJson)
	}
}

func (c *ProviderConfig) Validate() error {
	if c.apiTokens == nil || len(c.apiTokens) == 0 {
		return errors.New("no apiToken found in provider config")
	}
	if c.timeout < 0 {
		return errors.New("invalid timeout in config")
	}
	if c.protocol != protocolOpenAI && c.protocol != protocolOriginal {
		return errors.New("invalid protocol in config")
	}
	if c.context != nil {
		if err := c.context.Validate(); err != nil {
			return err
		}
	}

	if c.typ == "" {
		return errors.New("missing type in provider config")
	}
	initializer, has := providerInitializers[c.typ]
	if !has {
		return errors.New("unknown provider type: " + c.typ)
	}
	if err := initializer.ValidateConfig(*c); err != nil {
		return err
	}
	return nil
}

func (c *ProviderConfig) GetRandomToken() string {
	apiTokens := c.apiTokens
	count := len(apiTokens)
	switch count {
	case 0:
		return ""
	case 1:
		return apiTokens[0]
	default:
		return apiTokens[rand.Intn(count)]
	}
}

func CreateProvider(pc ProviderConfig) (Provider, error) {
	initializer, has := providerInitializers[pc.typ]
	if !has {
		return nil, errors.New("unknown provider type: " + pc.typ)
	}
	return initializer.CreateProvider(pc)
}

func getMappedModel(model string, modelMapping map[string]string, log wrapper.Log) string {
	if modelMapping == nil || len(modelMapping) == 0 {
		return model
	}
	if v, ok := modelMapping[model]; ok && len(v) != 0 {
		log.Debugf("model %s is mapped to %s explictly", model, v)
		return v
	}
	if v, ok := modelMapping[wildcard]; ok {
		log.Debugf("model %s is mapped to %s via wildcard", model, v)
		return v
	}
	return model
}
