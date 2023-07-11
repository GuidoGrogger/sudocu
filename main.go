package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"github.com/ServiceWeaver/weaver"
	"github.com/gorilla/mux"
)

func main() {
	if err := weaver.Run(context.Background(), serve); err != nil {
		log.Fatal(err)
	}
}

type app struct {
	weaver.Implements[weaver.Main]
	pdfGenerator      weaver.Ref[PDFGenerator]
	aDocRepository    weaver.Ref[ADocRepository]
	chatGPTRepository weaver.Ref[ChatGPTRepository]
	speechRepository  weaver.Ref[SpeechRepository]
	listener          weaver.Listener
}

func serve(ctx context.Context, a *app) error {
	logger := a.Logger()
	logger.Info("listener available on", a.listener)

	router := mux.NewRouter()

	router.HandleFunc("/pdf/{filename}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileName := vars["filename"]

		content, err := a.aDocRepository.Get().ReadFile(ctx, fileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			logger.Warn(err.Error())
			return
		}
		pdfContentBytes, err := a.pdfGenerator.Get().GeneratePDF(ctx, content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Warn(err.Error())
			return
		}

		// Serve the generated PDF
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "inline; filename=output.pdf")
		_, err = w.Write(pdfContentBytes)
		if err != nil {
			logger.Warn("Error writing response:", err)
		}
	})

	router.HandleFunc("/adoc/{filename}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileName := vars["filename"]

		content, err := a.aDocRepository.Get().ReadFile(ctx, fileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			logger.Warn(err.Error())
			return
		}

		// Serve the adoc data
		w.Header().Set("Content-Type", "text/plain")
		_, err = w.Write(content)
		if err != nil {
			logger.Warn("Error writing response:", err)
		}
	})

	router.HandleFunc("/iframe/{filename}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileName := vars["filename"]

		tmpl, err := template.ParseFiles("iframe.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"FileName": fileName,
		}

		w.Header().Set("Content-Type", "text/html")
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	router.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		adocFiles, err := a.aDocRepository.Get().GetFiles(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Warn(err.Error())
			return
		}

		type ViewData struct {
			ADocFiles []string
		}

		data := ViewData{
			ADocFiles: adocFiles,
		}

		// Parse the template file
		tmpl, err := template.ParseFiles("list.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Warn(err.Error())
			return
		}

		// Render the template with the provided data
		w.Header().Set("Content-Type", "text/html")
		err = tmpl.Execute(w, data)
		if err != nil {
			logger.Warn("Error writing response:", err)
		}
	})

	router.HandleFunc("/pdf/{filename}/change", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		vars := mux.Vars(r)
		fileName := vars["filename"]

		type RequestBody struct {
			Prompt string `json:"prompt"`
		}

		var requestBody RequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		oldMarkup, err := a.aDocRepository.Get().ReadFile(ctx, fileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Warn(err.Error())
			return
		}

		newMarkup, err := a.chatGPTRepository.Get().ChangeMarkup(ctx, string(oldMarkup), requestBody.Prompt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Warn(err.Error())
			return
		}

		err = a.aDocRepository.Get().SaveVariantForFile(ctx, fileName, newMarkup)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Warn(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	router.HandleFunc("/speech-to-text", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseMultipartForm(32 << 20) // Limit request size to 32MB
		if err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			logger.Warn(err.Error())
			return
		}

		file, _, err := r.FormFile("voicePrompt")
		if err != nil {
			http.Error(w, "Failed to retrieve voice prompt", http.StatusBadRequest)
			logger.Warn(err.Error())
			return
		}
		defer file.Close()

		audioBytes, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read audio file", http.StatusInternalServerError)
			logger.Warn(err.Error())
			return
		}

		// Call the text-to-speech method to convert audio bytes to text
		speechRepo := a.speechRepository.Get()
		text, err := speechRepo.SpeechToText(r.Context(), audioBytes)
		if err != nil {
			http.Error(w, "Failed to convert speech to text", http.StatusInternalServerError)
			logger.Warn(err.Error())
			return
		}
		a.Logger().Info("Received response from Whisper API: " + text)

		// Return the recognized text to the client
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(text))
	})

	http.Handle("/", router)

	return http.Serve(a.listener, nil)
}
