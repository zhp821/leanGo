package pojo

// Stock struct
type Stock struct {
	Code   string `json:"symbol"`
	Name   string `json:"name"`
	Time   string
	Pe     float32 `json:"dtsyl"`
	Price  float32 `json:"trade"`
	Avg5   float32
	Avg20  float32
	Avg60  float32
	Vavg5  float32
	Vavg20 float32
	K      float32
	D      float32
	J      float32
	Days   []Stock
	Open   float32
	Close  float32
	High   float32
	Low    float32
	Volume int
}
type SinaStock struct {
	Items []Stock
	Total int32
}
