package internal

type App struct {
	Config Config
}

func NewApp() App {
	config := LoadConfig()
	return App{Config: config}
}
