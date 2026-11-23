package mongo_kit

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// encryptionKeySize is the AES-256 key size in bytes.
	encryptionKeySize = 32

	// metaCollection holds mongo-kit's own documents in the client's database.
	metaCollection = "_mongokit"

	// encryptionCheckID is the _id of the key check document in metaCollection.
	encryptionCheckID = "encryption_check"

	// encryptionCheckPlaintext is the known value encrypted in the check
	// document. Never change it: databases already store it encrypted.
	encryptionCheckPlaintext = "mongo-kit encryption check v1"
)

// fieldCipher encrypts values with AES-256-GCM. The output is base64 of
// nonce + ciphertext + tag, so the same value encrypts differently each time.
type fieldCipher struct {
	gcm cipher.AEAD
}

func newFieldCipher(hexKey string) (*fieldCipher, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil || len(key) != encryptionKeySize {
		return nil, errors.New("must be 64 hex characters (32 bytes)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &fieldCipher{gcm: gcm}, nil
}

func (f *fieldCipher) encrypt(plaintext string) (string, error) {
	nonce := make([]byte, f.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("mongo: encryption failed: %w", err)
	}

	sealed := f.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (f *fieldCipher) decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("%w: not base64: %w", ErrDecryptionFailed, err)
	}

	nonceSize := f.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("%w: ciphertext too short", ErrDecryptionFailed)
	}

	plaintext, err := f.gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrDecryptionFailed, err)
	}

	return string(plaintext), nil
}

// EncryptionEnabled reports whether the client was created with WithEncryptionKey.
func (c *Client) EncryptionEnabled() bool {
	return c.cipher != nil
}

// Encrypt encrypts a value with the client's key (AES-256-GCM) and returns it
// base64-encoded, ready to store in a string field. Encrypting the same value
// twice gives different results, so encrypted fields cannot be queried by
// equality.
//
// Returns ErrEncryptionDisabled when the client has no encryption key.
//
// Example:
//
//	doc.Token, err = client.Encrypt(token)
func (c *Client) Encrypt(plaintext string) (string, error) {
	if c.cipher == nil {
		return "", ErrEncryptionDisabled
	}
	return c.cipher.encrypt(plaintext)
}

// Decrypt decrypts a value produced by Encrypt with the same key.
//
// Returns ErrEncryptionDisabled when the client has no encryption key, and an
// error wrapping ErrDecryptionFailed when the value is not valid ciphertext
// for this key.
//
// Example:
//
//	token, err := client.Decrypt(doc.Token)
func (c *Client) Decrypt(encoded string) (string, error) {
	if c.cipher == nil {
		return "", ErrEncryptionDisabled
	}
	return c.cipher.decrypt(encoded)
}

// verifyEncryptionKey makes sure every client of the database uses the same
// key. The first client stores encryptionCheckPlaintext encrypted with its key
// ($setOnInsert, so concurrent first starts keep a single value); every client
// then reads the stored value back and must decrypt it.
func (c *Client) verifyEncryptionKey(ctx context.Context) error {
	coll := c.defaultDB.Collection(metaCollection)

	value, err := c.cipher.encrypt(encryptionCheckPlaintext)
	if err != nil {
		return err
	}

	_, err = coll.UpdateOne(ctx,
		bson.M{"_id": encryptionCheckID},
		bson.M{"$setOnInsert": bson.M{"value": value, "created_at": time.Now().UTC()}},
		options.Update().SetUpsert(true),
	)
	// Two clients inserting at the same time: one wins, the other gets a
	// duplicate key error and only has to read what the winner stored.
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return newOperationError("encryption_check", err)
	}

	var stored struct {
		Value string `bson:"value"`
	}
	if err := coll.FindOne(ctx, bson.M{"_id": encryptionCheckID}).Decode(&stored); err != nil {
		return newOperationError("encryption_check", err)
	}

	plaintext, err := c.cipher.decrypt(stored.Value)
	if err != nil || plaintext != encryptionCheckPlaintext {
		return ErrEncryptionKeyMismatch
	}

	return nil
}
