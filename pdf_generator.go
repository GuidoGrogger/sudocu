package main

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/ServiceWeaver/weaver"
)

type PDFGenerator interface {
	GeneratePDF(context.Context, []byte) ([]byte, error)
}

// Implementation of the PDFGenerator component.
type pdfGenerator struct {
	weaver.Implements[PDFGenerator]
}

func (g *pdfGenerator) GeneratePDF(_ context.Context, content []byte) ([]byte, error) {
	// Generate PDF using AsciidoctorJ and a shell command
	cmd := exec.Command("asciidoctor-pdf", "-", "--theme", "default-sans")
	cmd.Stdin = bytes.NewReader(content)
	var pdfContent bytes.Buffer
	cmd.Stdout = &pdfContent

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return pdfContent.Bytes(), nil
}
