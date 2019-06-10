package config

import "sync"

// StaticProjectConfigManager maintains a static copy of the project config
type StaticProjectConfigManager struct {
	projectConfig ProjectConfig
	mutex         *sync.Mutex
}

// GetConfig returns the project config
func (configManager StaticProjectConfigManager) GetConfig() ProjectConfig {
	configManager.mutex.Lock()
	defer configManager.mutex.Unlock()
	return configManager.projectConfig
}

// SetConfig sets the project config
func (configManager *StaticProjectConfigManager) SetConfig(config ProjectConfig) {
	configManager.mutex.Lock()
	defer configManager.mutex.Unlock()
	configManager.projectConfig = config
}
