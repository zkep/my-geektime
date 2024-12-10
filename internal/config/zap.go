package config

type Zap struct {
	Directory     string `json:"directory"  yaml:"directory"`
	Prefix        string `json:"prefix" yaml:"prefix"`
	Format        string `json:"format" yaml:"format"`
	Level         string `json:"level" yaml:"level"`
	EncoderLevel  string `json:"encoder_level" yaml:"encoder_level"`
	StacktraceKey string `json:"stacktrace_key" yaml:"stacktrace_key"`
	MaxAge        int    `json:"max_age" yaml:"max_age"`
	ShowLine      bool   `json:"show_line" yaml:"show_line"`
	Console       bool   `json:"console" yaml:"console"`
}
