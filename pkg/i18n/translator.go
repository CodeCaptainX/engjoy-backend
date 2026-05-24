package i18n

var translations = map[string]map[string]string{
	"en": {
		"favorite_added":   "Favorite added",
		"favorite_removed": "Favorite removed",
		"favorites_fetched": "Favorites fetched",
		"sentence_created":  "Sentence created",
		"sentence_deleted":  "Sentence deleted",
		"sentence_analyzed": "Sentence analyzed",
		"environment_imported": "Environment pack imported",
	},
	"km": {
		"favorite_added":   "បានបន្ថែមទៅចំណូលចិត្ត",
		"favorite_removed": "បានដកចេញពីចំណូលចិត្ត",
		"favorites_fetched": "ទទួលបានចំណូលចិត្ត",
		"sentence_created":  "បានបង្កើតប្រយោគ",
		"sentence_deleted":  "បានលុបប្រយោគ",
		"sentence_analyzed": "បានវិភាគប្រយោគ",
		"environment_imported": "បាននាំចូលកញ្ចប់ទិន្នន័យ",
	},
}

func Translate(key, lang string) string {
	if langMap, ok := translations[lang]; ok {
		if val, ok := langMap[key]; ok {
			return val
		}
	}
	// Fallback to English
	if langMap, ok := translations["en"]; ok {
		if val, ok := langMap[key]; ok {
			return val
		}
	}
	return key
}
