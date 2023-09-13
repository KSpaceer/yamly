package generator

type EngineGenerator interface {
	DecodePackage() string
	EncodePackage() string
}

type EngineDecodeInfo struct {
	Package               string
	WarningSuppressorType string
}
