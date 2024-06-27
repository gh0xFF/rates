package model

// практика плохая, но для сервиса, который не изменяет данные сойдёт
// пока логика получения/сохранения несложная, можно оставить так

// по поводу CurrencyOfficialRate, хранить во float числа для работы с деньгами плохая практика
// оставил как есть, потому что я не выполняю никаких рассчётов и потери точности не будет

type RateModel struct {
	CurrencyId           uint32  `json:"Cur_ID" db:"currency_id"`                      //: 440
	CurrencyScale        uint32  `json:"Cur_Scale" db:"currency_scale"`                //: 1,
	CurrencyOfficialRate float32 `json:"Cur_OfficialRate" db:"currency_official_rate"` //: 2.1279
	CurrencyAbbrevation  string  `json:"Cur_Abbreviation" db:"-"`                      //: "AUD",
	CurrencyName         string  `json:"Cur_Name" db:"-"`                              //: "Австралийский доллар"
	Date                 string  `json:"Date" db:"date"`                               //: "2024-06-26T00:00:00"
}
