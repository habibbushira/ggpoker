package deck

import (
	"reflect"
	"testing"
)

func TestEncryptCard(t *testing.T) {
	key := []byte("MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDi1PmXlvXKG/qLfmlBFdBlrVzi1GNLcvZivCjGU/SnxUfBRAyVmWehjBLUksKvMbxxIp3iLLo5v8UT7BQSrxd6JAyeWghlo/zA5GpeJxejeohqhajwsoCIJ8PVzuyYM5ZyICQT5fxVPkn26AOR4smzjF8PFoo9JCxrimnVdxLkKQIDAQAB")
	card := Card{
		Suit:  Spades,
		Value: 1,
	}

	encOutput, err := EncryptCard(key, card)
	if err != nil {
		t.Errorf("Encription error: %s\n", err)
	}

	decCard, err := DecryptCard(key, encOutput)
	if err != nil {
		t.Errorf("Decryption error: %s\n", err)
	}

	if !reflect.DeepEqual(card, decCard) {
		t.Errorf("got %+v but expected %+v", decCard, card)
	}

}
