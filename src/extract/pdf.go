package extract

import (
	"fmt"

	"rsc.io/pdf"
)

func ExtractText(path string) error {
	r, err := pdf.Open(path)
	if err != nil {
		return err
	}

	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}

		content := p.Content()
		for _, txt := range content.Text {
			fmt.Print(txt.S)
		}
		fmt.Println("\n--- PAGE END ---")
	}

	return nil
}
