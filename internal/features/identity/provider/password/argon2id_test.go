package password

import (
	"strings"
	"testing"
)

func TestArgon2idHasherHashAndVerify(t *testing.T) {
	hasher := NewArgon2idHasher(DefaultArgon2idConfig())
	plainPassword := "correct horse battery staple"

	encodedHash, err := hasher.Hash(plainPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if strings.Contains(encodedHash, plainPassword) {
		t.Fatal("encoded hash must not contain the plain password")
	}

	matches, err := hasher.Verify(plainPassword, encodedHash)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if !matches {
		t.Fatal("expected password to match")
	}
}

func TestArgon2idHasherRejectsWrongPassword(t *testing.T) {
	hasher := NewArgon2idHasher(DefaultArgon2idConfig())
	encodedHash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	matches, err := hasher.Verify("wrong password", encodedHash)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if matches {
		t.Fatal("expected password mismatch")
	}
}

func TestArgon2idHasherUsesUniqueSalt(t *testing.T) {
	hasher := NewArgon2idHasher(DefaultArgon2idConfig())

	firstHash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password first time: %v", err)
	}
	secondHash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password second time: %v", err)
	}

	if firstHash == secondHash {
		t.Fatal("expected unique encoded hashes")
	}
}

func TestArgon2idHasherRejectsMalformedHash(t *testing.T) {
	hasher := NewArgon2idHasher(DefaultArgon2idConfig())

	_, err := hasher.Verify("correct horse battery staple", "not-an-argon2-hash")

	if err == nil {
		t.Fatal("expected malformed hash error")
	}
}
