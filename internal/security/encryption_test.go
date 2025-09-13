package security

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"
)

func TestAESGCMCipher_EncryptDecrypt(t *testing.T) {
	cipher := NewAESGCMCipher()

	// Generate a test key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	testData := []byte("Hello, World! This is a test message.")

	// Test encryption
	encrypted, err := cipher.Encrypt(testData, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if len(encrypted) <= len(testData) {
		t.Error("Encrypted data should be longer than original")
	}

	// Test decryption
	decrypted, err := cipher.Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(testData, decrypted) {
		t.Error("Decrypted data does not match original")
	}
}

func TestAESGCMCipher_InvalidKey(t *testing.T) {
	cipher := NewAESGCMCipher()
	testData := []byte("test data")

	// Test with invalid key length
	invalidKey := make([]byte, 16) // Should be 32 bytes

	_, err := cipher.Encrypt(testData, invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key length")
	}

	_, err = cipher.Decrypt(testData, invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key length")
	}
}

func TestAESGCMCipher_InvalidCiphertext(t *testing.T) {
	cipher := NewAESGCMCipher()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// Test with too short ciphertext
	shortCiphertext := []byte("short")

	_, err := cipher.Decrypt(shortCiphertext, key)
	if err == nil {
		t.Error("Expected error for too short ciphertext")
	}

	// Test with corrupted ciphertext
	validData := []byte("test data")
	encrypted, _ := cipher.Encrypt(validData, key)

	// Corrupt the ciphertext
	encrypted[len(encrypted)-1] ^= 0xFF

	_, err = cipher.Decrypt(encrypted, key)
	if err == nil {
		t.Error("Expected error for corrupted ciphertext")
	}
}

func TestKeyManager_GenerateKey(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	key, err := km.GetEncryptionKey()
	if err != nil {
		t.Fatalf("Failed to get encryption key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	// Test that we get the same key consistently
	key2, err := km.GetEncryptionKey()
	if err != nil {
		t.Fatalf("Failed to get encryption key second time: %v", err)
	}

	if !bytes.Equal(key, key2) {
		t.Error("Key manager should return consistent keys")
	}
}

func TestKeyManager_WithMasterKey(t *testing.T) {
	masterKey := make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}

	km, err := NewKeyManagerWithMasterKey(masterKey)
	if err != nil {
		t.Fatalf("Failed to create key manager with master key: %v", err)
	}

	key, err := km.GetEncryptionKey()
	if err != nil {
		t.Fatalf("Failed to get encryption key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	// Test with invalid master key length
	invalidMasterKey := make([]byte, 16)
	_, err = NewKeyManagerWithMasterKey(invalidMasterKey)
	if err == nil {
		t.Error("Expected error for invalid master key length")
	}
}

func TestKeyManager_DeriveKey(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	// Test deriving keys with different info strings
	key1, err := km.DeriveKey("test-purpose-1", 32)
	if err != nil {
		t.Fatalf("Failed to derive key 1: %v", err)
	}

	key2, err := km.DeriveKey("test-purpose-2", 32)
	if err != nil {
		t.Fatalf("Failed to derive key 2: %v", err)
	}

	if bytes.Equal(key1, key2) {
		t.Error("Different purposes should generate different keys")
	}

	// Test deriving the same key twice
	key3, err := km.DeriveKey("test-purpose-1", 32)
	if err != nil {
		t.Fatalf("Failed to derive key 3: %v", err)
	}

	if !bytes.Equal(key1, key3) {
		t.Error("Same purpose should generate same key")
	}

	// Test different key lengths
	shortKey, err := km.DeriveKey("test-short", 16)
	if err != nil {
		t.Fatalf("Failed to derive short key: %v", err)
	}

	if len(shortKey) != 16 {
		t.Errorf("Expected short key length 16, got %d", len(shortKey))
	}
}

func TestSecureStorage_StoreRetrieve(t *testing.T) {
	storage := &MockStorage{
		data: make(map[string]SecureCredential),
	}

	secureStorage, err := NewSecureStorage(storage)
	if err != nil {
		t.Fatalf("Failed to create secure storage: %v", err)
	}

	// Test storing credential
	testKey := "test-credential"
	testValue := "secret-value-123"

	err = secureStorage.StoreCredential(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to store credential: %v", err)
	}

	// Test retrieving credential
	retrieved, err := secureStorage.RetrieveCredential(testKey)
	if err != nil {
		t.Fatalf("Failed to retrieve credential: %v", err)
	}

	if retrieved != testValue {
		t.Errorf("Retrieved value %q does not match stored value %q", retrieved, testValue)
	}
}

func TestSecureStorage_RotateCredential(t *testing.T) {
	storage := &MockStorage{
		data: make(map[string]SecureCredential),
	}

	secureStorage, err := NewSecureStorage(storage)
	if err != nil {
		t.Fatalf("Failed to create secure storage: %v", err)
	}

	testKey := "test-credential"
	originalValue := "original-secret"
	newValue := "new-secret"

	// Store original credential
	err = secureStorage.StoreCredential(testKey, originalValue)
	if err != nil {
		t.Fatalf("Failed to store original credential: %v", err)
	}

	// Get original credential to check version
	originalCred, err := storage.Load(testKey)
	if err != nil {
		t.Fatalf("Failed to load original credential: %v", err)
	}

	// Rotate credential
	err = secureStorage.RotateCredential(testKey, newValue)
	if err != nil {
		t.Fatalf("Failed to rotate credential: %v", err)
	}

	// Verify new value
	retrieved, err := secureStorage.RetrieveCredential(testKey)
	if err != nil {
		t.Fatalf("Failed to retrieve rotated credential: %v", err)
	}

	if retrieved != newValue {
		t.Errorf("Retrieved rotated value %q does not match expected %q", retrieved, newValue)
	}

	// Verify version was incremented
	rotatedCred, err := storage.Load(testKey)
	if err != nil {
		t.Fatalf("Failed to load rotated credential: %v", err)
	}

	if rotatedCred.Version != originalCred.Version+1 {
		t.Errorf("Expected version %d, got %d", originalCred.Version+1, rotatedCred.Version)
	}

	// Verify creation time is preserved
	if !rotatedCred.CreatedAt.Equal(originalCred.CreatedAt) {
		t.Error("Creation time should be preserved during rotation")
	}
}

func TestSecureStorage_DeleteCredential(t *testing.T) {
	storage := &MockStorage{
		data: make(map[string]SecureCredential),
	}

	secureStorage, err := NewSecureStorage(storage)
	if err != nil {
		t.Fatalf("Failed to create secure storage: %v", err)
	}

	testKey := "test-credential"
	testValue := "secret-value"

	// Store credential
	err = secureStorage.StoreCredential(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to store credential: %v", err)
	}

	// Delete credential
	err = secureStorage.DeleteCredential(testKey)
	if err != nil {
		t.Fatalf("Failed to delete credential: %v", err)
	}

	// Verify deletion
	_, err = secureStorage.RetrieveCredential(testKey)
	if err == nil {
		t.Error("Expected error when retrieving deleted credential")
	}
}

func TestSecureStorage_ListCredentials(t *testing.T) {
	storage := &MockStorage{
		data: make(map[string]SecureCredential),
	}

	secureStorage, err := NewSecureStorage(storage)
	if err != nil {
		t.Fatalf("Failed to create secure storage: %v", err)
	}

	// Store multiple credentials
	credentials := map[string]string{
		"cred1": "value1",
		"cred2": "value2",
		"cred3": "value3",
	}

	for key, value := range credentials {
		err = secureStorage.StoreCredential(key, value)
		if err != nil {
			t.Fatalf("Failed to store credential %s: %v", key, err)
		}
	}

	// List credentials
	list, err := secureStorage.ListCredentials()
	if err != nil {
		t.Fatalf("Failed to list credentials: %v", err)
	}

	if len(list) != len(credentials) {
		t.Errorf("Expected %d credentials, got %d", len(credentials), len(list))
	}

	// Verify all credentials are present
	found := make(map[string]bool)
	for _, cred := range list {
		found[cred.Key] = true
	}

	for key := range credentials {
		if !found[key] {
			t.Errorf("Credential %s not found in list", key)
		}
	}
}

func TestSecureStorage_WithMasterKey(t *testing.T) {
	storage := &MockStorage{
		data: make(map[string]SecureCredential),
	}

	masterKey := make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}

	secureStorage, err := NewSecureStorageWithMasterKey(storage, masterKey)
	if err != nil {
		t.Fatalf("Failed to create secure storage with master key: %v", err)
	}

	testKey := "test-credential"
	testValue := "secret-value"

	// Store and retrieve credential
	err = secureStorage.StoreCredential(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to store credential: %v", err)
	}

	retrieved, err := secureStorage.RetrieveCredential(testKey)
	if err != nil {
		t.Fatalf("Failed to retrieve credential: %v", err)
	}

	if retrieved != testValue {
		t.Errorf("Retrieved value does not match stored value")
	}

	// Verify master key is accessible
	retrievedMasterKey := secureStorage.GetMasterKey()
	if !bytes.Equal(masterKey, retrievedMasterKey) {
		t.Error("Retrieved master key does not match original")
	}
}

// MockStorage implements Storage interface for testing
type MockStorage struct {
	data map[string]SecureCredential
}

func (ms *MockStorage) Save(key string, credential SecureCredential) error {
	ms.data[key] = credential
	return nil
}

func (ms *MockStorage) Load(key string) (*SecureCredential, error) {
	if cred, exists := ms.data[key]; exists {
		return &cred, nil
	}
	return nil, ErrCredentialNotFound
}

func (ms *MockStorage) Delete(key string) error {
	if _, exists := ms.data[key]; !exists {
		return ErrCredentialNotFound
	}
	delete(ms.data, key)
	return nil
}

func (ms *MockStorage) List() ([]string, error) {
	keys := make([]string, 0, len(ms.data))
	for key := range ms.data {
		keys = append(keys, key)
	}
	return keys, nil
}

func TestSecureCredential_Timestamps(t *testing.T) {
	storage := &MockStorage{
		data: make(map[string]SecureCredential),
	}

	secureStorage, err := NewSecureStorage(storage)
	if err != nil {
		t.Fatalf("Failed to create secure storage: %v", err)
	}

	testKey := "test-credential"
	testValue := "secret-value"

	beforeStore := time.Now()

	// Store credential
	err = secureStorage.StoreCredential(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to store credential: %v", err)
	}

	afterStore := time.Now()

	// Load credential directly to check timestamps
	cred, err := storage.Load(testKey)
	if err != nil {
		t.Fatalf("Failed to load credential: %v", err)
	}

	// Check creation time
	if cred.CreatedAt.Before(beforeStore) || cred.CreatedAt.After(afterStore) {
		t.Error("Creation time is not within expected range")
	}

	// Check last used time
	if cred.LastUsedAt.Before(beforeStore) || cred.LastUsedAt.After(afterStore) {
		t.Error("Last used time is not within expected range")
	}

	// Retrieve credential (should update last used time)
	time.Sleep(10 * time.Millisecond) // Small delay to ensure timestamp difference
	beforeRetrieve := time.Now()

	_, err = secureStorage.RetrieveCredential(testKey)
	if err != nil {
		t.Fatalf("Failed to retrieve credential: %v", err)
	}

	afterRetrieve := time.Now()

	// Load credential again to check updated timestamp
	updatedCred, err := storage.Load(testKey)
	if err != nil {
		t.Fatalf("Failed to load updated credential: %v", err)
	}

	// Last used time should be updated
	if !updatedCred.LastUsedAt.After(cred.LastUsedAt) {
		t.Error("Last used time should be updated after retrieval")
	}

	if updatedCred.LastUsedAt.Before(beforeRetrieve) || updatedCred.LastUsedAt.After(afterRetrieve) {
		t.Error("Updated last used time is not within expected range")
	}

	// Creation time should remain unchanged
	if !updatedCred.CreatedAt.Equal(cred.CreatedAt) {
		t.Error("Creation time should not change during retrieval")
	}
}

func BenchmarkAESGCMCipher_Encrypt(b *testing.B) {
	cipher := NewAESGCMCipher()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		b.Fatalf("Failed to generate key: %v", err)
	}
	data := make([]byte, 1024) // 1KB test data
	if _, err := rand.Read(data); err != nil {
		b.Fatalf("Failed to generate test data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cipher.Encrypt(data, key)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}
	}
}

func BenchmarkAESGCMCipher_Decrypt(b *testing.B) {
	cipher := NewAESGCMCipher()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		b.Fatalf("Failed to generate key: %v", err)
	}
	data := make([]byte, 1024) // 1KB test data
	if _, err := rand.Read(data); err != nil {
		b.Fatalf("Failed to generate test data: %v", err)
	}

	encrypted, err := cipher.Encrypt(data, key)
	if err != nil {
		b.Fatalf("Failed to encrypt test data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cipher.Decrypt(encrypted, key)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}
	}
}

func BenchmarkSecureStorage_StoreRetrieve(b *testing.B) {
	storage := &MockStorage{
		data: make(map[string]SecureCredential),
	}

	secureStorage, err := NewSecureStorage(storage)
	if err != nil {
		b.Fatalf("Failed to create secure storage: %v", err)
	}

	testValue := "secret-value-for-benchmarking"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := string(rune(i))

		err := secureStorage.StoreCredential(key, testValue)
		if err != nil {
			b.Fatalf("Failed to store credential: %v", err)
		}

		_, err = secureStorage.RetrieveCredential(key)
		if err != nil {
			b.Fatalf("Failed to retrieve credential: %v", err)
		}
	}
}
