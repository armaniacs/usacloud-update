package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/hkdf"
)

// Cipher defines the interface for encryption/decryption operations
type Cipher interface {
	Encrypt(plaintext []byte, key []byte) ([]byte, error)
	Decrypt(ciphertext []byte, key []byte) ([]byte, error)
}

// AESGCMCipher implements AES-GCM encryption
type AESGCMCipher struct{}

// NewAESGCMCipher creates a new AES-GCM cipher
func NewAESGCMCipher() *AESGCMCipher {
	return &AESGCMCipher{}
}

// Encrypt encrypts plaintext using AES-GCM
func (c *AESGCMCipher) Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Ensure key is 32 bytes for AES-256
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes for AES-256, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-GCM
func (c *AESGCMCipher) Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Ensure key is 32 bytes for AES-256
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes for AES-256, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// KeyManager handles encryption key management
type KeyManager struct {
	masterKey []byte
}

// NewKeyManager creates a new key manager
func NewKeyManager() (*KeyManager, error) {
	// Generate a random master key
	masterKey := make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	return &KeyManager{
		masterKey: masterKey,
	}, nil
}

// NewKeyManagerWithMasterKey creates a key manager with existing master key
func NewKeyManagerWithMasterKey(masterKey []byte) (*KeyManager, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes, got %d", len(masterKey))
	}

	return &KeyManager{
		masterKey: masterKey,
	}, nil
}

// GetEncryptionKey derives an encryption key using HKDF
func (km *KeyManager) GetEncryptionKey() ([]byte, error) {
	return km.DeriveKey("credential-encryption", 32)
}

// DeriveKey derives a key for specific purpose using HKDF
func (km *KeyManager) DeriveKey(info string, length int) ([]byte, error) {
	hkdf := hkdf.New(sha256.New, km.masterKey, nil, []byte(info))

	key := make([]byte, length)
	if _, err := io.ReadFull(hkdf, key); err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	return key, nil
}

// Storage defines the interface for credential storage
type Storage interface {
	Save(key string, credential SecureCredential) error
	Load(key string) (*SecureCredential, error)
	Delete(key string) error
	List() ([]string, error)
}

// SecureCredential represents an encrypted credential
type SecureCredential struct {
	Key        string    `json:"key"`
	Encrypted  []byte    `json:"encrypted"`
	Algorithm  string    `json:"algorithm"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	Version    int       `json:"version"`
}

// SecureStorage provides secure credential storage
type SecureStorage struct {
	cipher     Cipher
	keyManager *KeyManager
	storage    Storage
}

// NewSecureStorage creates a new secure storage instance
func NewSecureStorage(storage Storage) (*SecureStorage, error) {
	keyManager, err := NewKeyManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create key manager: %w", err)
	}

	return &SecureStorage{
		cipher:     NewAESGCMCipher(),
		keyManager: keyManager,
		storage:    storage,
	}, nil
}

// NewSecureStorageWithMasterKey creates secure storage with existing master key
func NewSecureStorageWithMasterKey(storage Storage, masterKey []byte) (*SecureStorage, error) {
	keyManager, err := NewKeyManagerWithMasterKey(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create key manager: %w", err)
	}

	return &SecureStorage{
		cipher:     NewAESGCMCipher(),
		keyManager: keyManager,
		storage:    storage,
	}, nil
}

// StoreCredential encrypts and stores a credential
func (ss *SecureStorage) StoreCredential(key, value string) error {
	// Get encryption key
	encKey, err := ss.keyManager.GetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Encrypt the credential value
	encrypted, err := ss.cipher.Encrypt([]byte(value), encKey)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Create secure credential
	credential := SecureCredential{
		Key:        key,
		Encrypted:  encrypted,
		Algorithm:  "AES-256-GCM",
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		Version:    1,
	}

	// Store encrypted credential
	return ss.storage.Save(key, credential)
}

// RetrieveCredential decrypts and returns a credential
func (ss *SecureStorage) RetrieveCredential(key string) (string, error) {
	// Load encrypted credential
	credential, err := ss.storage.Load(key)
	if err != nil {
		return "", fmt.Errorf("failed to load credential: %w", err)
	}

	// Get decryption key
	encKey, err := ss.keyManager.GetEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Decrypt the credential
	plaintext, err := ss.cipher.Decrypt(credential.Encrypted, encKey)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	// Update last used time
	credential.LastUsedAt = time.Now()
	if saveErr := ss.storage.Save(key, *credential); saveErr != nil {
		// Log save error but continue with decryption result
	}

	return string(plaintext), nil
}

// DeleteCredential removes a credential
func (ss *SecureStorage) DeleteCredential(key string) error {
	return ss.storage.Delete(key)
}

// ListCredentials returns all stored credential metadata
func (ss *SecureStorage) ListCredentials() ([]*SecureCredential, error) {
	keys, err := ss.storage.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}

	credentials := make([]*SecureCredential, 0, len(keys))
	for _, key := range keys {
		cred, err := ss.storage.Load(key)
		if err != nil {
			continue // Skip corrupted entries
		}
		credentials = append(credentials, cred)
	}

	return credentials, nil
}

// RotateCredential updates an existing credential with a new value
func (ss *SecureStorage) RotateCredential(key, newValue string) error {
	// Load existing credential to get metadata
	existing, err := ss.storage.Load(key)
	if err != nil {
		return fmt.Errorf("failed to load existing credential: %w", err)
	}

	// Get encryption key
	encKey, err := ss.keyManager.GetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Encrypt new value
	encrypted, err := ss.cipher.Encrypt([]byte(newValue), encKey)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Update credential with new encrypted value and incremented version
	updated := SecureCredential{
		Key:        key,
		Encrypted:  encrypted,
		Algorithm:  "AES-256-GCM",
		CreatedAt:  existing.CreatedAt,
		LastUsedAt: time.Now(),
		Version:    existing.Version + 1,
	}

	return ss.storage.Save(key, updated)
}

// GetMasterKey returns the master key (for backup/restore purposes)
func (ss *SecureStorage) GetMasterKey() []byte {
	key := make([]byte, len(ss.keyManager.masterKey))
	copy(key, ss.keyManager.masterKey)
	return key
}

// FileStorage implements Storage interface using files
type FileStorage struct {
	basePath string
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(basePath string) *FileStorage {
	return &FileStorage{
		basePath: basePath,
	}
}

// Save stores a credential to file
func (fs *FileStorage) Save(key string, credential SecureCredential) error {
	filename := fmt.Sprintf("%s.json", key)
	filepath := fmt.Sprintf("%s/%s", fs.basePath, filename)

	data, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("failed to marshal credential: %w", err)
	}

	return writeFileSecure(filepath, data)
}

// Load loads a credential from file
func (fs *FileStorage) Load(key string) (*SecureCredential, error) {
	filename := fmt.Sprintf("%s.json", key)
	filepath := fmt.Sprintf("%s/%s", fs.basePath, filename)

	data, err := readFileSecure(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credential file: %w", err)
	}

	var credential SecureCredential
	if err := json.Unmarshal(data, &credential); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential: %w", err)
	}

	return &credential, nil
}

// Delete removes a credential file
func (fs *FileStorage) Delete(key string) error {
	filename := fmt.Sprintf("%s.json", key)
	filepath := fmt.Sprintf("%s/%s", fs.basePath, filename)

	return secureDelete(filepath)
}

// List returns all credential keys
func (fs *FileStorage) List() ([]string, error) {
	// This would be implemented to scan the directory
	// For now, return empty list
	return []string{}, nil
}

// Helper functions for secure file operations
func writeFileSecure(filepath string, data []byte) error {
	// Implementation would ensure proper file permissions (0600)
	// and atomic writes
	return fmt.Errorf("file operations not implemented in this example")
}

func readFileSecure(filepath string) ([]byte, error) {
	// Implementation would ensure proper error handling
	// and secure file reading
	return nil, fmt.Errorf("file operations not implemented in this example")
}

func secureDelete(filepath string) error {
	// Implementation would perform secure deletion
	// (overwrite with random data before unlinking)
	return fmt.Errorf("secure delete not implemented in this example")
}
