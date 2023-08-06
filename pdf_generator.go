package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"text/template"

	"github.com/ServiceWeaver/weaver"
)

type PDFGenerator interface {
	GeneratePDF(_ context.Context, ascii_doc []byte, json []byte) ([]byte, error)
}

// Implementation of the PDFGenerator component.
type pdfGenerator struct {
	weaver.Implements[PDFGenerator]
}

func (g *pdfGenerator) GeneratePDF(_ context.Context, ascii_doc []byte, json_bytes []byte) ([]byte, error) {

	var jsonData map[string]interface{}
	json.Unmarshal([]byte(json_bytes), &jsonData)

	// Parse the template
	tpl, err := template.New("").Parse(string(ascii_doc))
	if err != nil {
		return nil, err
	}

	// Use a buffer to capture the result
	var buf bytes.Buffer

	// Execute the template
	if err := tpl.Execute(&buf, jsonData); err != nil {
		return nil, err
	}

	// Generate PDF using AsciidoctorJ and a shell command
	cmd := exec.Command("asciidoctor-pdf", "-", "--theme", "default-sans")
	cmd.Stdin = bytes.NewReader(buf.Bytes())
	var pdfContent bytes.Buffer
	cmd.Stdout = &pdfContent

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return pdfContent.Bytes(), nil
}
