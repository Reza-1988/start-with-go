package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetAllPhoneNumbersSample(p *Phonebook) []string {
	contacts := p.GetAllContacts()
	phones := make([]string, 0, len(contacts))
	for _, contact := range contacts {
		p, err := p.GetPhoneNumber(contact)
		if err != nil {
			panic(err)
		}
		phones = append(phones, p)
	}
	return phones
}
func PhonebookOfSample(args ...string) *Phonebook {
	if len(args)%2 != 0 {
		panic("cannot make phone-book")
	}
	ph := NewPhonebook()
	for i := 0; i < len(args)/2; i++ {
		status := ph.AddContact(args[i*2], args[i*2+1])
		if status != "Phone Number Added" {
			panic("bad contact to add")
		}
	}
	return ph
}
func assertSamePhoneBookSample(t *testing.T, m1, m2 *Phonebook) {
	assert.ElementsMatch(t, m1.GetAllContacts(), m2.GetAllContacts())
	assert.ElementsMatch(t, GetAllPhoneNumbersSample(m1), GetAllPhoneNumbersSample(m2))
}
func TestSample(t *testing.T) {
	ph := NewPhonebook()
	status := ph.AddContact("ali", "09546765543")
	assert.Equal(t, "Phone Number Added", status)
	status = ph.AddContact("taghi", "09123333333")
	assert.Equal(t, "Phone Number Added", status)
	status = ph.AddContact("ali", "09121111111")
	assert.Equal(t, "Phone Number Updated", status)
	status = ph.AddContact("zari", "09123333333")
	assert.Equal(t, "Phone Number Added", status)

	got := ph.GetAllContacts()
	assert.ElementsMatch(t, []string{"ali", "taghi", "zari"}, got)

	number, err := ph.GetPhoneNumber("ali")
	assert.Nil(t, err)
	assert.Equal(t, number, "09121111111")

	ph.DeleteDuplicatePhoneNumbers()
	assertSamePhoneBookSample(t, ph, PhonebookOfSample(
		"ali", "09121111111",
		"taghi", "09123333333",
	))

	ph.AddContact("nazi", "987876543")
	ph.AddContact("niki", "0932567876")
	ph.DeleteInvalidPhoneNumbers()
	assertSamePhoneBookSample(t, ph, PhonebookOfSample(
		"ali", "09121111111",
		"taghi", "09123333333",
	))

	var s []string = ph.FindCatchyPhoneNumbers()
	assert.Equal(t, 2, len(s))
	assert.Equal(t, "09121111111", s[0])
	assert.Equal(t, "09123333333", s[1])
}
