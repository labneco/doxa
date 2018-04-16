// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

// A LangDirection defines the direction of a language
// either left-to-right or right-to-left
type LangDirection string

const (
	// LangDirectionLTR defines a language written from left to right
	LangDirectionLTR LangDirection = "ltr"
	// LangDirectionRTL defines a language written from right to left
	LangDirectionRTL LangDirection = "rtl"
)

// LangParameters defines the parameters of a language locale
type LangParameters struct {
	DateFormat   string        `json:"date_format"`
	Direction    LangDirection `json:"lang_direction"`
	ThousandsSep string        `json:"thousands_sep"`
	TimeFormat   string        `json:"time_format"`
	DecimalPoint string        `json:"decimal_point"`
	ID           int64         `json:"id"`
	Grouping     string        `json:"grouping"`
}
