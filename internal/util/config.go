package util

import "github.com/mxcd/go-config/config"

func InitConfig() error {
	err := config.LoadConfig([]config.Value{
		config.String("LOG_LEVEL").NotEmpty().Default("info"),
		config.Int("PORT").Default(8080),

		config.Bool("DEV").Default(false),

		config.String("GOTENBERG_URL").NotEmpty().Default("http://localhost:3000"),
	})
	return err
}
