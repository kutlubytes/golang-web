package main

import (
	"time"
	"fmt"
	"encoding/json"
	"encoding/base64"
	"crypto/hmac"
	"crypto/sha256"
	"strings"
)

var secretkey string

func main() {
	secretkey = "mysecretkey"
	payload := map[string]interface{} {
		"name": "ahmet",
		"admin": 1,
		"exp": time.Now().Add(time.Minute).Unix(),
	}

	// username && password control
	token := generateToken(payload)
	isValid := validateToken(token)
	fmt.Println(token)
	fmt.Println(isValid)
}

func generateToken(payload map[string]interface{}) string{
	header := map[string]string {
		"alg": "HS256",
		"typ": "JWT",
	}

	// Convert to JSON
	header_json, _ := json.Marshal(header)
	payload_json, _ := json.Marshal(payload)

	// Encode to base64
	header_b64 := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(header_json)
	payload_b64 := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(payload_json)
	
	// Concat header and payload
	data := fmt.Sprintf("%v.%v", header_b64, payload_b64)
	
	// Make signature
	sign := hmac.New(sha256.New, []byte(secretkey))
	sign.Write([]byte(data))
	signature := sign.Sum(nil)

	// Encode signature
	signature_b64 := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(signature)

	token := fmt.Sprintf("%v.%v.%v", header_b64, payload_b64, signature_b64)
	return token
}

func validateToken(token string) bool{
	//  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6MSwiZXhwIjoxNjE2NDg2NTk0LCJuYW1lIjoiYWhtZXQifQ.h7SPeBLlZsSIVqwFZpITGImbusghEixpEvWG025Imkc
	parts := strings.Split(token, ".")
	header_b64 := parts[0]
	payload_b64 := parts[1]
	signature_b64 := parts[2] // From user which gets from auth server.

	// Decode header
	header_json, _ := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(header_b64)

	// Unmarshal header
	var header map[string]string
	json.Unmarshal(header_json, &header)
	alg := header["alg"]
	typ := header["typ"]
	if typ != "JWT" {
		fmt.Println("JWT")
		return false
	}

	if alg != "HS256" {
		fmt.Println("HS256")
		return false
	}

	// Verify signature
		// Concat header and payload
	data := fmt.Sprintf("%v.%v", header_b64, payload_b64)
	
		// Make signature
	sign := hmac.New(sha256.New, []byte(secretkey))
	sign.Write([]byte(data))
	signature := sign.Sum(nil) // with header, payload and secret key.
	signature_check_b64 := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(signature)

	if string(signature_check_b64) != string(signature_b64) {
		fmt.Println("compare")
		return false
	}

	// Check exp.
		// Decode payload
	payload_json, _ := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(payload_b64)
	var payload map[string]interface{}
	json.Unmarshal(payload_json, &payload)

	exp := int64(payload["exp"].(float64))
	now := time.Now().Unix()

	if exp < now {
		fmt.Println("expired")
		return false
	}

	return true
}
