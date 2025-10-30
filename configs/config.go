package configs

import "github.com/spf13/viper" 

type Config struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	ApiPort    string `mapstructure:"API_PORT"`
	JWTSecret  string `mapstructure:"JWT_SECRET_KEY"`
}

func LoadConfig(path string) (config *Config, err error) {
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("API_PORT")
	viper.BindEnv("JWT_SECRET_KEY")


	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.ReadInConfig()

	err = viper.Unmarshal(&config)
	return
}
