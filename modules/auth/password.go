package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory      uint32 = 64 * 1024
	argonIterations  uint32 = 3
	argonParallelism uint8  = 2
	argonSaltLength         = 16
	argonKeyLength          = 32
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)
	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonIterations,
		argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func VerifyPassword(password, encoded string) bool {
	params, salt, expected, err := decodeHash(encoded)
	if err != nil {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.parallelism, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

type argonParams struct {
	memory      uint32
	time        uint32
	parallelism uint8
}

func decodeHash(encoded string) (argonParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "argon2id" || parts[1] != "v=19" {
		return argonParams{}, nil, nil, fmt.Errorf("invalid password hash")
	}
	paramParts := strings.Split(parts[2], ",")
	params := argonParams{}
	for _, part := range paramParts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return argonParams{}, nil, nil, fmt.Errorf("invalid argon parameter")
		}
		n, err := strconv.ParseUint(kv[1], 10, 32)
		if err != nil {
			return argonParams{}, nil, nil, err
		}
		switch kv[0] {
		case "m":
			params.memory = uint32(n)
		case "t":
			params.time = uint32(n)
		case "p":
			params.parallelism = uint8(n)
		}
	}
	if params.memory == 0 || params.time == 0 || params.parallelism == 0 {
		return argonParams{}, nil, nil, fmt.Errorf("missing argon parameter")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return argonParams{}, nil, nil, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return argonParams{}, nil, nil, err
	}
	return params, salt, hash, nil
}
