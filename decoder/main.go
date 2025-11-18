package main

func StartDecipher(senderChan chan string, decipherer func(encrypted string) string) chan string {

	outputChan := make(chan string, 5)

	go func() {
		for encrypted := range senderChan {
			decrypted := decipherer(encrypted)
			outputChan <- decrypted
		}
	}()
	return outputChan
}
