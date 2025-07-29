package gorote

import (
	"testing"
)

func TestValidatePassword(t *testing.T) {
	t.Run("valida senha com maiúscula e símbolo", func(t *testing.T) {
		err := ValidatePassword("Senha@123")
		if err != nil {
			t.Errorf("esperava senha válida, mas retornou erro: %v", err)
		}
	})

	t.Run("sem maiúscula", func(t *testing.T) {
		err := ValidatePassword("senha@123")
		if err == nil || err.Error() != "uppercase-password must contain at least one uppercase letter" {
			t.Errorf("esperava erro de maiúscula, mas retornou: %v", err)
		}
	})

	t.Run("sem símbolo", func(t *testing.T) {
		err := ValidatePassword("Senha123")
		if err == nil || err.Error() != "symbol-password must contain at least one symbol" {
			t.Errorf("esperava erro de símbolo, mas retornou: %v", err)
		}
	})
}

func TestHashPasswordAndCheck(t *testing.T) {
	password := "Senha@123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("erro ao hashear senha: %v", err)
	}
	if !CheckPasswordHash(password, hash) {
		t.Error("senha não bate com o hash gerado")
	}
	if CheckPasswordHash("outraSenha", hash) {
		t.Error("hash deveria falhar para senha errada")
	}
}
