package model

type InvestorType struct {
	Code     string `db:"code"`
	LabelEn  string `db:"label_en"`
	LabelID  string `db:"label_id"`
	Category string `db:"category"`
}
