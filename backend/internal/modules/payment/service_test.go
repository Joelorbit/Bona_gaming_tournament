package payment

import (
	"testing"

	"bona-backend/internal/repository"
)

func TestCustomerNameUsesDisplayName(t *testing.T) {
	displayName := "Abebe Kebede"
	first, last := customerName(repository.Profile{
		Username:    "user_12345678",
		DisplayName: &displayName,
	})

	if first != "Abebe" || last != "Kebede" {
		t.Fatalf("customerName() = %q %q, want Abebe Kebede", first, last)
	}
}

func TestCustomerNameSanitizesInvalidUsername(t *testing.T) {
	first, last := customerName(repository.Profile{
		Username: "user_12345678",
	})

	if first != "User" || last != "Player" {
		t.Fatalf("customerName() = %q %q, want User Player", first, last)
	}
}

func TestCustomerNameFallsBackToEmailTokens(t *testing.T) {
	email := "abebe.kebede92@example.com"
	first, last := customerName(repository.Profile{
		Username: "x_1",
		Email:    &email,
	})

	if first != "Abebe" || last != "Kebede" {
		t.Fatalf("customerName() = %q %q, want Abebe Kebede", first, last)
	}
}

func TestCustomerNameDropsShortAndNumericTokens(t *testing.T) {
	displayName := "A 123 B"
	first, last := customerName(repository.Profile{
		Username:    "p9",
		DisplayName: &displayName,
	})

	if first != "Bona" || last != "Player" {
		t.Fatalf("customerName() = %q %q, want Bona Player", first, last)
	}
}
