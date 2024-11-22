package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type Session struct {
	UserId string `json:"userId"`
}

func (server *GameServer) createSession(session *Session) (string, error) {
	macKeyStr := server.config.MacKey
	macKey, err := base64.StdEncoding.DecodeString(macKeyStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode configured mac key: %v", err)
	}

	sessionDataStr, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session data: %v", err)
	}
	mac := hmac.New(sha256.New, macKey)
	mac.Write(sessionDataStr)
	expectedMAC := mac.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(sessionDataStr) + ":" + base64.RawURLEncoding.EncodeToString(expectedMAC), nil
}

func (server *GameServer) validateSession(sessionStr string) (*Session, error) {
	parts := strings.SplitN(sessionStr, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid session: %s", sessionStr)
	}

	actualMac, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid session (2): %s", sessionStr)
	}
	sessionDataStr, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid session (3): %s", sessionStr)
	}

	macKeyStr := server.config.MacKey
	macKey, err := base64.StdEncoding.DecodeString(macKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode configured mac key: %v", err)
	}

	mac := hmac.New(sha256.New, macKey)
	mac.Write(sessionDataStr)
	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(actualMac, expectedMAC) {
		return nil, fmt.Errorf("invalid session (4): %s", sessionStr)
	}

	sessionData := new(Session)
	err = json.Unmarshal(sessionDataStr, sessionData)
	if err != nil {
		return nil, fmt.Errorf("invalid session (5): %s", sessionStr)
	}

	return sessionData, nil
}
