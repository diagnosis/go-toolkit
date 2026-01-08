package secure

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	password := "StopTradingFocusStudying123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !match {
		t.Fatal("password should match")
	}

	//verify wrong password
	wrong, err := VerifyPassword("PowerNeedsToBeControlled", hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if wrong {
		t.Fatal("wrong password match")
	}

}
