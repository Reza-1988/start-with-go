package main

type Phonebook struct {
	Contacts map[string]string
}

func NewPhonebook() *Phonebook {
	return &Phonebook{
		make(map[string]string),
	}
}

func (p *Phonebook) GetPhoneNumber(contactName string) (string, error) {
	return "", nil
}

func (p *Phonebook) AddContact(contactName string, phoneNumber string) string {
	return ""
}

func (p *Phonebook) GetAllContacts() []string {
	return nil
}

func (p *Phonebook) GetAllContactsStartsWith(letter rune) []string {
	return nil
}

func (p *Phonebook) DeleteInvalidPhoneNumbers() {
}

func (p *Phonebook) DeleteDuplicatePhoneNumbers() {
}

func (p *Phonebook) FindCatchyPhoneNumbers() []string {
	return nil
}
