package langfuse

import (
	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"github.com/xyzbit/ino/config"
)

var flusher func()

func Init() {
	langfuseConfig := config.AppConfig.Langfuse
	if !langfuseConfig.Enabled {
		return
	}

	var cbh *langfuse.CallbackHandler
	cbh, flusher = langfuse.NewLangfuseHandler(&langfuse.Config{
		Host:      langfuseConfig.Host,
		PublicKey: langfuseConfig.PublicKey,
		SecretKey: langfuseConfig.SecretKey,
	})

	callbacks.AppendGlobalHandlers(cbh) // 设置langfuse为全局callback
}

func Close() {
	if flusher != nil {
		flusher()
	}
}
