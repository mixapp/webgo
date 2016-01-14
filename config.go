package webgo

var CFG Config

type Config struct {
	Author string
}
func (cfg *Config) Set () {
	// TODO: Запись значения в файл настроек + установка нового значения для текущего экземпляра
}

func init() {
	CFG = Config{"Nikita"}
	// TODO: Реализовать чтение настроек из файла config в папке приложения, формат YAML?
}
