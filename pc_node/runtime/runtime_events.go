package runtime

type RuntimeEvents struct {
	OnNodeStarted   func()
	OnNodeStopped   func()
	OnModuleLoaded  func(moduleName string)
	OnModuleFailed  func(moduleName string, err error)
	OnHealthChanged func(moduleName string, status string)
	OnShutdown      func()
	OnRestart       func()
}
