package main

import (
	"bytes"
	"context"
	"strings"

	"github.com/ServiceWeaver/weaver"
	"github.com/bytesparadise/libasciidoc"
	"github.com/bytesparadise/libasciidoc/pkg/configuration"
	"github.com/jung-kurt/gofpdf"
)

type PDFGenerator interface {
	GeneratePDF(context.Context, []byte) ([]byte, error)
}

// Implementation of the PDFGenerator component.
type pdfGenerator struct {
	weaver.Implements[PDFGenerator]
}

func (g *pdfGenerator) GeneratePDF(_ context.Context, content []byte) ([]byte, error) {
	reader := strings.NewReader(string(content))
	var buf bytes.Buffer

	config := configuration.NewConfiguration(
		configuration.WithBackEnd("html"))

	g.Logger().Error("Trying to convert asciidoc to html ", string(content))

	_, err := libasciidoc.Convert(reader, &buf, config)
	if err != nil {
		g.Logger().Error("Error converting AsciiDoc to HTML5: ", err.Error())
		return nil, err
	}

	// Log the generated PDF.
	g.Logger().Info("Generated HTML: ", string(buf.Bytes()))

	// convert html to a pdf and return it as a byte slice
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	html := pdf.HTMLBasicNew()
	html.Write(5, buf.String())

	// create a buffer to hold pdf bytes
	var pdfBuf bytes.Buffer
	err = pdf.Output(&pdfBuf)
	if err != nil {
		g.Logger().Error("Error converting HTML to PDF: ", err.Error())
		return nil, err
	}

	return pdfBuf.Bytes(), nil
}
