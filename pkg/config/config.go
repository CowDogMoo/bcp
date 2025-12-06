/*
Copyright © 2025 Jayson Grace <jayson.e.grace@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package config

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/cowdogmoo/bcp/pkg/logging"
	"github.com/cowdogmoo/bcp/pkg/model"
	"github.com/spf13/viper"
)

var (
	// GlobalConfig holds the global configuration
	GlobalConfig model.Config

	// MaxRetries is the maximum number of retry attempts for AWS operations
	MaxRetries = 3

	// RetryDelay is the base delay in seconds between retries
	RetryDelay = 2
)

// Init initializes the configuration system
func Init(cfgFile string) error {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in home directory and current directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(home, ".bcp"))
		viper.AddConfigPath("/etc/bcp")
	}

	// Set defaults
	setDefaults()

	// Read environment variables
	viper.SetEnvPrefix("BCP")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug("No config file found, using defaults")
		} else {
			return fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		log.Debug("Using config file: %s", viper.ConfigFileUsed())
	}

	// Unmarshal config
	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load constants
	LoadConstants()

	// Initialize logging
	log.Init(GlobalConfig.Log.Format, GlobalConfig.Log.Level)

	return nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Log defaults
	viper.SetDefault("log.format", "text")
	viper.SetDefault("log.level", "info")

	// AWS defaults
	viper.SetDefault("aws.region", "us-east-1")
	viper.SetDefault("aws.profile", "default")
	viper.SetDefault("aws.bucket", "")

	// Transfer defaults
	viper.SetDefault("transfer.max_retries", 3)
	viper.SetDefault("transfer.retry_delay", 2)
}

// LoadConstants loads configuration constants
func LoadConstants() {
	MaxRetries = viper.GetInt("transfer.max_retries")
	if MaxRetries == 0 {
		MaxRetries = 3
	}

	RetryDelay = viper.GetInt("transfer.retry_delay")
	if RetryDelay == 0 {
		RetryDelay = 2
	}
}

// GetBucket returns the configured S3 bucket name
func GetBucket() string {
	return GlobalConfig.AWS.Bucket
}

// GetRegion returns the configured AWS region
func GetRegion() string {
	return GlobalConfig.AWS.Region
}

// GetProfile returns the configured AWS profile
func GetProfile() string {
	return GlobalConfig.AWS.Profile
}
