package serve

import (
	"net"

	"go.bbkane.com/shovel/dig"
)

type TdData struct {
	Content string
	Rowspan int
}
type AnsErrCount struct {
	AnsErrs []string
	Count   int
}

type Row struct {
	Columns      []TdData
	AnsErrCounts []AnsErrCount
}
type Table struct {
	FilledFormURL string
	Rows          []Row
}

type buildTableParams struct {
	Qnames        []string
	RtypeStrs     []string
	Subnets       []net.IP
	Nameservers   []string
	ResMul        []dig.DigRepeatResult
	SubnetToName  map[string]string
	FilledFormURL string
}

func buildTable(p buildTableParams) Table {
	// Add params to output table
	qLen := len(p.Qnames)
	rLen := len(p.RtypeStrs)
	sLen := len(p.Subnets)
	nLen := len(p.Nameservers)
	rows := qLen * rLen * sLen * nLen
	t := Table{
		FilledFormURL: p.FilledFormURL,
		Rows:          make([]Row, rows),
	}

	qWidth := rows / qLen
	{
		i := 0
		for r := 0; r < rows; r += qWidth {
			td := TdData{Content: p.Qnames[i%qLen], Rowspan: qWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	rWidth := qWidth / rLen
	{
		i := 0
		for r := 0; r < rows; r += rWidth {
			td := TdData{Content: p.RtypeStrs[i%rLen], Rowspan: rWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	sWidth := rWidth / sLen
	{
		i := 0
		for r := 0; r < rows; r += sWidth {
			subnet := p.Subnets[i%sLen]
			content := subnet.String()
			if subnet != nil {
				content = content + " (" + p.SubnetToName[content] + ")"
			}
			td := TdData{Content: content, Rowspan: sWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	nWidth := sWidth / nLen
	{
		i := 0
		for r := 0; r < rows; r += nWidth {
			td := TdData{Content: p.Nameservers[i%nLen], Rowspan: nWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	// Add anserrs to table
	for i, r := range p.ResMul {
		aecs := []AnsErrCount{}
		for _, a := range r.Answers {
			aecs = append(
				aecs,
				AnsErrCount{AnsErrs: a.StringSlice, Count: a.Count},
			)
		}
		for _, e := range r.Errors {
			aecs = append(
				aecs,
				AnsErrCount{AnsErrs: []string{e.String}, Count: e.Count},
			)
		}
		t.Rows[i].AnsErrCounts = aecs

	}

	return t
}
