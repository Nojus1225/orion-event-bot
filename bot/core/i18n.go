package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var Bundle *i18n.Bundle

func InitI18n() {
	Bundle = i18n.NewBundle(language.English)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	entries, err := os.ReadDir("locales")
	if err != nil {
		fmt.Println("Warning: could not read locales dir:", err)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		locale := entry.Name()
		dir := filepath.Join("locales", locale)
		files, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println("Warning: could not read dir:", dir, err)
			continue
		}
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
				if file.Name() == "commands.json" {
					continue
				}
				path := filepath.Join(dir, file.Name())
				buf, err := os.ReadFile(path)
				if err != nil {
					fmt.Println("Warning: could not read", path, ":", err)
					continue
				}
				msgFile, err := Bundle.ParseMessageFileBytes(buf, file.Name())
				if err != nil {
					fmt.Println("Warning: could not parse", path, ":", err)
					continue
				}
				tag := language.MustParse(locale)
				err = Bundle.AddMessages(tag, msgFile.Messages...)
				if err != nil {
					fmt.Println("Warning: could not add messages from", path, ":", err)
				}
			}
		}
	}
}

func GetLocalizer(lang string) *i18n.Localizer {
	if Bundle == nil {
		InitI18n()
	}
	return i18n.NewLocalizer(Bundle, lang, "en-US")
}
