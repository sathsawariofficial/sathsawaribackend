package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"rideshare/pkgs/configuration"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var secret []byte

func init() {
	secret = []byte(configuration.ConfigurationData.Auth.Secrect)
}

// Secret key for signing the token (use a secure key in production)

// CreateJWT generates a JWT token with a driver_id claim
func CreateJWT(sessionId string, tokenClaims map[string]string) (string, error) {
	logger.LogInfo("Request received in CreateJWT", sessionId)

	// Define the token expiration time
	expirationTime := time.Now().Add(time.Duration(configuration.ConfigurationData.Auth.ExpirationTime) * time.Second)

	// Create the claims
	claims := jwt.MapClaims{
		"exp": expirationTime.Unix(),
		"iat": time.Now().Unix(),
	}

	for key, val := range tokenClaims {
		claims[key] = val
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	logger.LogInfo("Response returned from CreateJWT", sessionId)

	// Sign the token with the secret key
	return token.SignedString(secret)
}

// VerifyJWT validates a JWT token and extracts the driver_id
func VerifyJWT(sessionId, tokenString string) (string, error) {
	logger.LogInfo("Request received in VerifyJWT", sessionId)

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		logger.LogError(sessionId, err)
		// Handle specific error for expired token
		if err.Error() == "token is expired" {
			return "", fmt.Errorf("token has expired")
		}

		return "", fmt.Errorf("error parsing token: %w", err)
	}

	// Extract and verify claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, ok := claims["tokenType"].(string)
		if !ok {
			err = fmt.Errorf("token type not found")
			logger.LogError(sessionId, err)
			return "", err
		}

		if tokenType == constants.DRIVER_TOKEN {
			driverID, ok := claims["driverId"].(string)
			if !ok {
				err = fmt.Errorf("driverId not found or invalid")
				logger.LogError(sessionId, err)
				return "", err
			}

			return driverID, nil
		} else if tokenType == constants.ADMIN_TOKEN {
			adminID, ok := claims["adminId"].(string)
			if !ok {
				err = fmt.Errorf("adminId not found or invalid")
				logger.LogError(sessionId, err)
				return "", err
			}

			return adminID, nil
		} else if tokenType == constants.OPEN_TOKEN {
			return "", nil
		} else {
			return "", fmt.Errorf("invalid token type")
		}
	}

	logger.LogInfo("Response returned from VerifyJWT", sessionId)

	err = fmt.Errorf("invalid token")
	logger.LogError(sessionId, err)

	return "", err
}

func HashPassword(sessionId, password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		logger.LogError(sessionId, err)
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

func ComparePassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(plainPassword),
	)
}

// Encrypt encrypts the plaintext using AES with the provided key
func EncryptAES(sessionId, plaintext string) (string, error) {
	logger.LogInfo("Request received in EncryptAES", sessionId)

	// Convert key and plaintext to byte slices
	keyBytes := []byte(configuration.ConfigurationData.Auth.IV16)
	plaintextBytes := []byte(plaintext)

	// Ensure key is 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
	if len(keyBytes) != 16 && len(keyBytes) != 24 && len(keyBytes) != 32 {
		err := fmt.Errorf("key must be 16, 24, or 32 bytes long")
		logger.LogError(sessionId, err)
		return "", err
	}

	// Create AES block cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		logger.LogError(sessionId, err)
		return "", fmt.Errorf("error creating cipher: %w", err)
	}

	// Generate a random initialization vector (IV)
	ciphertext := make([]byte, aes.BlockSize+len(plaintextBytes))
	// Use a fixed IV (insecure but deterministic)
	iv := []byte(configuration.ConfigurationData.Auth.IV16) // 16 bytes
	if len(iv) != aes.BlockSize {
		err = errors.New("IV must be 16 bytes long")
		logger.LogError(sessionId, err)
		return "", err
	}

	// Encrypt the plaintext
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintextBytes)

	logger.LogInfo("Response returned from EncryptAES", sessionId)

	// Return the base64-encoded ciphertext
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the base64-encoded ciphertext using AES with the provided key
func DecryptAES(sessionId, encrypted string) (string, error) {
	logger.LogInfo("Request received in DecryptAES", sessionId)

	// Convert key to byte slice
	keyBytes := []byte(configuration.ConfigurationData.Auth.IV16)

	// Ensure key is 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
	if len(keyBytes) != 16 && len(keyBytes) != 24 && len(keyBytes) != 32 {
		err := fmt.Errorf("key must be 16, 24, or 32 bytes long")
		logger.LogError(sessionId, err)
		return "", err
	}

	// Decode the base64-encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		logger.LogError(sessionId, err)
		return "", fmt.Errorf("error decoding ciphertext: %w", err)
	}

	// Create AES block cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		logger.LogError(sessionId, err)
		return "", fmt.Errorf("error creating cipher: %w", err)
	}

	// Extract the IV
	iv := []byte(configuration.ConfigurationData.Auth.IV16) // 16 bytes
	if len(iv) != aes.BlockSize {
		err = errors.New("IV must be 16 bytes long")
		logger.LogError(sessionId, err)
		return "", err
	}
	ciphertext = ciphertext[aes.BlockSize:]

	// Decrypt the ciphertext
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	logger.LogInfo("Response returned from DecryptAES", sessionId)

	// Return the decrypted plaintext as a string
	return string(ciphertext), nil
}
