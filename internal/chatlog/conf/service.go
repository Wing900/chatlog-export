package conf

import (
	"log"
	"os"
	"sync"

	"github.com/sjzar/chatlog/pkg/config"
)

const (
	ConfigName   = "chatlog"
	ConfigType   = "json"
	EnvConfigDir = "CHATLOG_DIR"
)

// Service 配置服务
type Service struct {
	configPath string
	config     *Config
	mu         sync.RWMutex
}

// NewService 创建配置服务
func NewService(configPath string) (*Service, error) {

	service := &Service{
		configPath: configPath,
	}

	if err := service.Load(); err != nil {
		return nil, err
	}

	return service, nil
}

// Load 加载配置
func (s *Service) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	configPath := s.configPath
	if configPath == "" {
		configPath = os.Getenv(EnvConfigDir)
	}
	if err := config.Init(ConfigName, ConfigType, configPath); err != nil {
		log.Fatal(err)
	}

	conf := &Config{}
	if err := config.Load(conf); err != nil {
		log.Fatal(err)
	}
	conf.ConfigDir = config.ConfigPath
	s.config = conf
	return nil
}

// GetConfig 获取配置副本
func (s *Service) GetConfig() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回配置副本
	configCopy := *s.config
	return &configCopy
}
