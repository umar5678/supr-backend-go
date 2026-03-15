package config

import (
    "time"
)

type KafkaConfig struct {
    Brokers []string      `mapstructure:"brokers"`
    SASL    SAslConfig    `mapstructure:"sasl"`
    Producer ProducerConfig `mapstructure:"producer"`
    Consumer ConsumerConfig `mapstructure:"consumer"`
}

type SAslConfig struct {
    Enabled   bool   `mapstructure:"enabled"`
    Mechanism string `mapstructure:"mechanism"`
    Username  string `mapstructure:"username"`
    Password  string `mapstructure:"password"`
}

type ProducerConfig struct {
    Retries       int           `mapstructure:"retries"`
    BatchSize     int           `mapstructure:"batch_size"`
    FlushInterval time.Duration `mapstructure:"flush_interval"`
    Timeout       time.Duration `mapstructure:"timeout"`
    Compression   string        `mapstructure:"compression"`
}

type ConsumerConfig struct {
    GroupID              string        `mapstructure:"group_id"`
    MaxConcurrent        int           `mapstructure:"max_concurrent"`
    SessionTimeout       time.Duration `mapstructure:"session_timeout"`
    HeartbeatInterval    time.Duration `mapstructure:"heartbeat_interval"`
    CommitInterval       time.Duration `mapstructure:"commit_interval"`
    MaxRetries           int           `mapstructure:"max_retries"`
    RetryBackoff         time.Duration `mapstructure:"retry_backoff"`
}

type NotificationConfig struct {
    SendGrid SendGridConfig `mapstructure:"sendgrid"`
    Twilio   TwilioConfig   `mapstructure:"twilio"`
    Firebase FirebaseConfig `mapstructure:"firebase"`
}

type SendGridConfig struct {
    APIKey    string `mapstructure:"api_key"`
    FromEmail string `mapstructure:"from_email"`
    FromName  string `mapstructure:"from_name"`
}

type TwilioConfig struct {
    AccountSID string `mapstructure:"account_sid"`
    AuthToken  string `mapstructure:"auth_token"`
    FromNumber string `mapstructure:"from_number"`
}

type FirebaseConfig struct {
    CredentialsFile string `mapstructure:"credentials_file"`
}
