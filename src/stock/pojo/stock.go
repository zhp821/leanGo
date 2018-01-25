package pojo

// Stock struct
type Stock struct {
	Code   string `json:"symbol"`
	Name   string `json:"name"`
	Time   string
	Pe     float32 `json:"dtsyl"`
	Price  float32 `json:"trade"`
	Avg5   float32
	Avg10  float32
	Avg20  float32
	Avg60  float32
	Vavg5  float32
	Vavg20 float32
	K      []float32
	D      []float32
	J      []float32
	K60    []float32
	D60    []float32
	J60    []float32
	K30    []float32
	D30    []float32
	J30    []float32
	K15    []float32
	D15    []float32
	J15    []float32
	Days   []Stock
	Open   float32
	Close  float32
	High   float32
	Low    float32
	Volume int
	M30    []Stock
	M60    []Stock
	Num    int
	Max10  float32
	Min10  float32
	Max20  float32
	Min20  float32
}
type View struct {
	Code  string
	Name  string
	Time  string
	Price float32
	Open  float32
	Avg5  float32
	Avg20 float32
	Des   string
	Url   string
}

type SinaStock struct {
	Items []Stock
	Total int32
}

var Tmpl = `
<html>
<body>
<h1>{{.total}} stocks</h1>
<table>
<tr style='text-align: left'>
  <th>title</th>
  <th>code</th>
  <th>Price</th>
  <th>kdj</th>
  <th>des</th>
</tr>
{{range .stocks}}
<tr>
  <td>{{.Name}}</td>
  <td><span onclick="copyToClipboard('{{.Code}}')">{{.Code}}</span></td>
  <td>{{.Price}}</a></td>
  <td><a  href="{{.Url}}" target="_blank">{{.Avg5}}-{{.Avg20}}</a></td>
  <td>{{.Des}}</td>
</tr>
{{end}}
</table>
<script>
 function copyToClipboard(content) {
      var aux = document.createElement("input");
      aux.setAttribute("value", content);    
      document.body.appendChild(aux);
      aux.select();
      document.execCommand("copy");
      document.body.removeChild(aux);
    }
</script>
</body>
</html>
`
