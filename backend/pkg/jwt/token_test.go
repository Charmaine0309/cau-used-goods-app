package jwt

import "testing"

func TestTokenTypesAndVersion(t *testing.T) {
	access, err := Generate("secret", 1, 12, "USER", 3)
	if err != nil {
		t.Fatal(err)
	}
	accessClaims, err := Parse("secret", access)
	if err != nil {
		t.Fatal(err)
	}
	if accessClaims.TokenType != TokenTypeAccess || accessClaims.TokenVersion != 3 {
		t.Fatalf("unexpected access claims: %#v", accessClaims)
	}

	reactivation, err := GenerateReactivation("secret", 12, 4)
	if err != nil {
		t.Fatal(err)
	}
	reactivationClaims, err := Parse("secret", reactivation)
	if err != nil {
		t.Fatal(err)
	}
	if reactivationClaims.TokenType != TokenTypeReactivate || reactivationClaims.TokenVersion != 4 {
		t.Fatalf("unexpected reactivation claims: %#v", reactivationClaims)
	}
}
