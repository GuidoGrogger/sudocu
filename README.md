# Project Name: Sudocu

This project, Sudocu, is a prototype that aims to explore the combination of GPT from OpenAI, Whisper for speech-to-text translation, and document generation capabilities. The goal is to provide a web-based interface where users can change the content of a PDF document by sending a speech command through their browser.

## Key Features

- Utilizes GPT from OpenAI for document generation based on user prompts.
- Employs the Whisper API for translating speech to text.
- Enables users to modify the content of a PDF document by sending a speech command.
- The server processes the received text and updates the AsciiDoc file accordingly.
- The AsciiDoc file is transformed into a PDF and presented to the user.

## Usage

1. Set your OpenAI API key by executing the command: `export OPENAI_API_KEY=<KEY>`
2. Run the following command: `weaver generate . && SERVICEWEAVER_CONFIG=weaver.toml go run .`
3. Open your web browser and navigate to http://localhost:8080/list.
4. Choose a document from the list (you can create your own documents in the /adoc folder).
5. Click and hold the "Voice" button, then speak the desired change to be made.
6. Edit the text of the change as needed and press "Send."
7. The server processes the text, updates the AsciiDoc file based on the prompt, and generates a new PDF.
8. The modified PDF is displayed to the user.

Please note that this prototype relies on the combination of GPT, Whisper, and document generation, and may have limitations or areas for improvement. It is designed to showcase the integration of these technologies and provide an interactive experience for users to experiment with changing document content through speech commands.

## Disclaimer

Please be aware that this prototype is provided "as is" and is not intended for production use. It is an experimental project for exploring the capabilities of GPT, Whisper, and document generation. Use it at your own risk.

For more information or support, please refer to the project's documentation or reach out to the project maintainers.