package security

import "errors"

// Security package errors
var (
	// Encryption errors
	ErrInvalidKeyLength    = errors.New("invalid key length")
	ErrEncryptionFailed    = errors.New("encryption failed")
	ErrDecryptionFailed    = errors.New("decryption failed")
	ErrKeyDerivationFailed = errors.New("key derivation failed")
	ErrInvalidCiphertext   = errors.New("invalid ciphertext")

	// Storage errors
	ErrCredentialNotFound = errors.New("credential not found")
	ErrCredentialExists   = errors.New("credential already exists")
	ErrStorageOperation   = errors.New("storage operation failed")
	ErrInvalidCredential  = errors.New("invalid credential")

	// Pattern errors
	ErrInvalidPatternName  = errors.New("invalid pattern name")
	ErrInvalidPatternRegex = errors.New("invalid pattern regex")
	ErrPatternNotFound     = errors.New("pattern not found")

	// Input errors
	ErrEmptyInput   = errors.New("empty input")
	ErrInvalidInput = errors.New("invalid input")
	ErrInputTooLong = errors.New("input too long")

	// Audit errors
	ErrAuditLogFailed    = errors.New("audit log failed")
	ErrInvalidAuditEvent = errors.New("invalid audit event")

	// Authentication errors
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrSessionExpired       = errors.New("session expired")
)
