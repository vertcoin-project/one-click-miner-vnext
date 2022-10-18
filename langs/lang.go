package langs

type Lang interface {
	GetCode() string
	GetName() string
}

func GetLangs(lang string) []Lang {
	return []Lang{
		// add all languages
	}
}

func GetLang(lang string) Lang {
	langs := GetLangs(lang)
	for _, p := range langs {
		if p.GetCode() == lang {
			return p
		}
	}
	return langs[0]
}
