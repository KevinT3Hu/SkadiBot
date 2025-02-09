package utils

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type ProbConfig struct {
	HitProb        float64 `yaml:"hit_prob"`
	MissProb       float64 `yaml:"miss_prob"`
	AIFeedProb     float64 `yaml:"ai_feed_prob"`
	AIResponseProb float64 `yaml:"ai_response_prob"`
}

type PromptConfig struct {
	AIRequestPrompt   string `yaml:"ai_request_prompt"`
	AIAtRequestPrompt string `yaml:"ai_at_request_prompt"`
}

type AIApiConfig struct {
	BaseUrl string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
}

type Doc2VecConfig struct {
	Enable      bool   `yaml:"enable"`
	Destination string `yaml:"destination"`
}

type Config struct {
	ProbConfig    ProbConfig    `yaml:"prob_config"`
	PromptConfig  PromptConfig  `yaml:"prompt_config"`
	AIApiConfig   AIApiConfig   `yaml:"ai_api_config"`
	Doc2VecConfig Doc2VecConfig `yaml:"doc2vec_config"`
}

func writeConfigToFile(config Config) error {
	configContent, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	configFile := "/etc/skadi/config.yaml"
	err = os.WriteFile(configFile, configContent, 0644)
	if err != nil {
		return err
	}
	return nil
}

func configInit() Config {
	// try to load config from /etc/skadi/config.yaml
	// if not found, use default config
	configFile := "/etc/skadi/config.yaml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		c := Config{
			ProbConfig: ProbConfig{
				HitProb:        0.2,
				MissProb:       0.15,
				AIFeedProb:     0.9,
				AIResponseProb: 0.3,
			},
		}
		writeConfigToFile(c)
		return c
	}

	// load config from file
	configContent, err := os.ReadFile(configFile)
	if err != nil {
		SLogger.Warn("Failed to load config file", "err", err)
		return Config{}
	}
	config := Config{}
	err = yaml.Unmarshal(configContent, &config)
	if err != nil {
		SLogger.Warn("Failed to parse config file", "err", err)
		return Config{}
	}
	return config
}

var (
	config               Config = configInit()
	configMu             sync.Mutex
	ProbGeneratorManager = NewProbGeneratorManager(config.ProbConfig)
	AIChatterClient      = NewAiChatter(config.AIApiConfig)
	Doc2VecAddrChan      = make(chan string)
)

func GetConfig() Config {
	configMu.Lock()
	defer configMu.Unlock()
	return config
}

func UpdateConfig(newConf Config) {
	configMu.Lock()
	defer configMu.Unlock()
	config = newConf
	writeConfigToFile(config)
}
