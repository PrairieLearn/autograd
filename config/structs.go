package config

type Config struct {
	AMQP       AMQPConfig       `yaml:"amqp"`
	GraderRepo GraderRepoConfig `yaml:"grader_repo"`
}

type AMQPConfig struct {
	URL          string `yaml:"url"`
	GradingQueue string `yaml:"grading_queue"`
	ResultQueue  string `yaml:"result_queue"`
}

type GraderRepoConfig struct {
	RepoURL     string     `yaml:"repo_url"`
	Commit      string     `yaml:"commit"`
	Credentials CredConfig `yaml:"credentials"`
}

type CredConfig struct {
	PublicKey  string `yaml:"public_key"`
	PrivateKey string `yaml:"private_key"`
	Passphrase string `yaml:"passphrase"`
}
