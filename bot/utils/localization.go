package utils

import "github.com/nicksnyder/go-i18n/v2/i18n"

func LocalizeText(loc *i18n.Localizer, id string, data map[string]interface{}, fallback string) string {
	msg, err := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
	})
	if err != nil || msg == "" {
		return fallback
	}
	return msg
}
