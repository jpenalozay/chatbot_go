// chatbot/utils/validateSend.go

package utils

import (
	"bytes"
	"chatbot/logger"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

// ValidateSignature verifica la firma HMAC-SHA256 de una solicitud.
func ValidateSignature(payload, signature, secret string) bool {
	logger.Log.Info("Validating signature")

	// Crear un nuevo HMAC con SHA256
	mac := hmac.New(sha256.New, []byte(secret))

	// Escribir el payload en el HMAC
	mac.Write([]byte(payload))

	// Obtener la firma esperada en formato hexadecimal
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	// Comparar la firma esperada con la firma proporcionada
	isValid := hmac.Equal([]byte(signature), []byte(expectedMAC))

	if isValid {
		logger.Log.Info("Signature is valid")
	} else {
		logger.Log.Warn("Invalid signature detected")
	}

	return isValid
}

// getTextMessageInput prepara los datos en formato JSON para enviar mensajes de texto a través de WhatsApp.
func getTextMessageInput(recipient, text string) map[string]interface{} {
	logger.Log.Info("Preparing text message input")

	return map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                recipient,
		"type":              "text",
		"text": map[string]string{
			"preview_url": "false",
			"body":        text,
		},
	}
}

// SendMessage envía un mensaje a través de la API de WhatsApp
func SendMessage(phone string, messageData map[string]interface{}) error {
	logger.Log.Info("Sending message to WhatsApp API")

	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", os.Getenv("VERSION"), os.Getenv("PHONE_NUMBER_ID"))

	reqBody, err := json.Marshal(messageData)
	if err != nil {
		logger.Log.Errorf("Failed to marshal request body: %v", err)
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	logger.Log.Infof("Preparing to send message. URL: %s, Body: %s", url, string(reqBody))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Log.Errorf("Failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("ACCESS_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Errorf("Failed to send message: %v", err)
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	logger.Log.Infof("Response status code: %d", resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Errorf("Failed to read response body: %v", err)
		return fmt.Errorf("failed to read response body: %w", err)
	}
	logger.Log.Infof("Response body: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		logger.Log.Errorf("Failed to send message, status code: %d", resp.StatusCode)
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	logger.Log.Info("Message sent successfully")
	return nil
}

// ProcessTextForWhatsApp formatea el texto para WhatsApp.
func ProcessTextForWhatsApp(text string) string {
	logger.Log.Info("Processing text for WhatsApp")

	// Crear una expresión regular para encontrar el texto entre ** **
	re := regexp.MustCompile(`\*\*(.*?)\*\*`)

	// Reemplazar el texto entre ** ** por * *
	processed := re.ReplaceAllString(text, "*$1*")

	logger.Log.Info("Text processed successfully")
	return processed
}

// IsValidWhatsAppMessage valida la estructura del mensaje de WhatsApp.
func IsValidWhatsAppMessage(body map[string]interface{}) bool {
	logger.Log.Info("Validating WhatsApp message structure")

	// Verificar si el mensaje tiene la estructura esperada
	_, ok := body["entry"].([]interface{})[0].(map[string]interface{})["changes"].([]interface{})[0].(map[string]interface{})["value"].(map[string]interface{})["messages"]

	if ok {
		logger.Log.Info("Valid WhatsApp message structure")
	} else {
		logger.Log.Warn("Invalid WhatsApp message structure")
	}

	return ok
}

// IsWhatsAppStatusUpdate verifica si el mensaje es una actualización de estado de WhatsApp.
func IsWhatsAppStatusUpdate(body map[string]interface{}) bool {
	_, ok := body["entry"].([]interface{})[0].(map[string]interface{})["changes"].([]interface{})[0].(map[string]interface{})["value"].(map[string]interface{})["statuses"]
	return ok
}

// GetTextMessageInput prepara los datos en formato JSON para enviar mensajes de texto a través de WhatsApp.
func GetTextMessageInput(recipient, text string) map[string]interface{} {
	return map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                recipient,
		"type":              "text",
		"text": map[string]string{
			"preview_url": "false",
			"body":        text,
		},
	}
}

// GetInteractiveMessageInput prepara los datos en formato JSON para enviar mensajes interactivos con botones.
func GetInteractiveMessageInput(recipient, text string, buttons []string) map[string]interface{} {
	buttonObjects := make([]map[string]interface{}, len(buttons))
	for i, button := range buttons {
		buttonObjects[i] = map[string]interface{}{
			"type": "reply",
			"reply": map[string]string{
				"id":    fmt.Sprintf("button_%d", i+1),
				"title": button,
			},
		}
	}

	return map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                recipient,
		"type":              "interactive",
		"interactive": map[string]interface{}{
			"type": "button",
			"body": map[string]string{
				"text": text,
			},
			"action": map[string]interface{}{
				"buttons": buttonObjects,
			},
		},
	}
}
