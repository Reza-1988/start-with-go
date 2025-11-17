package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Phonebook struct {
	Contacts map[string]string
}

func NewPhonebook() *Phonebook {
	return &Phonebook{
		make(map[string]string),
	}
}

func (p *Phonebook) GetPhoneNumber(contactName string) (string, error) {
	number, ok := p.Contacts[contactName]
	if !ok {
		return "", fmt.Errorf("%s not exists", contactName)
	}
	return number, nil
}

func (p *Phonebook) AddContact(contactName string, phoneNumber string) string {
	number, ok := p.Contacts[contactName]
	if ok {
		if number == phoneNumber {
			return "No Change"
		}
		p.Contacts[contactName] = phoneNumber
		return "Phone Number Updated"

	}
	p.Contacts[contactName] = phoneNumber
	return "Phone Number Added"
}

func (p *Phonebook) GetAllContacts() []string {
	contactNames := make([]string, 0)
	for name := range p.Contacts {
		contactNames = append(contactNames, name)
	}
	return contactNames
}

func (p *Phonebook) GetAllContactsStartsWith(letter rune) []string {
	nameStartWith := make([]string, 0)
	for name := range p.Contacts {
		if strings.HasPrefix(name, string(letter)) {
			nameStartWith = append(nameStartWith, name)
		}
	}
	return nameStartWith
}

func (p *Phonebook) DeleteInvalidPhoneNumbers() {
	for name, number := range p.Contacts {
		if len(number) != 11 {
			delete(p.Contacts, name)
			continue
		}
		if !strings.HasPrefix(number, "09") && !strings.HasPrefix(number, "+9") {
			delete(p.Contacts, name)
			continue
		}
		for _, r := range number[1:] {
			if !unicode.IsDigit(r) {
				delete(p.Contacts, name)
				break
			}
		}
	}
}

func (p *Phonebook) DeleteDuplicatePhoneNumbers() {
	phoneToBestName := make(map[string]string)
	for name, number := range p.Contacts {
		if strings.HasPrefix(number, "+9") {
			number = "0" + number[1:]
		}
		bestName, ok := phoneToBestName[number]
		if !ok {
			phoneToBestName[number] = name
		} else {
			if name < bestName {
				phoneToBestName[number] = name
			}
		}
	}
	for name, number := range p.Contacts {
		if strings.HasPrefix(number, "+9") {
			number = "0" + number[1:]
		}
		if name != phoneToBestName[number] {
			delete(p.Contacts, name)
		}
	}
}

func (p *Phonebook) FindCatchyPhoneNumbers() []string {

	type phoneToScore struct {
		phone string
		score int
	}

	scores := make([]phoneToScore, 0, len(p.Contacts))

	for _, number := range p.Contacts {

		score := Score(number[2:])
		scores = append(scores, phoneToScore{
			phone: number,
			score: score,
		})
	}

	sort.Slice(scores, func(i, j int) bool {
		if scores[i].score == scores[j].score {
			return scores[i].phone < scores[j].phone
		}
		return scores[i].score > scores[j].score
	})

	if len(scores) <= 3 {
		res := make([]string, 0, len(scores))
		for _, p := range scores {
			res = append(res, p.phone)
		}
		return res
	}

	edge := scores[2].score
	limit := 3
	for limit < len(scores) && scores[limit].score == edge {
		limit++
	}

	res := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		res = append(res, scores[i].phone)
	}
	return res
}

func Score(number string) int {
	score := 0
	n := len(number)

	for i := 0; i <= n-3; i++ {
		if number[i] == number[i+1] && number[i+1] == number[i+2] {
			score += 10
		}
	}

	for i := 0; i <= n-4; i++ {
		a, _ := strconv.Atoi(number[i : i+2])
		b, _ := strconv.Atoi(number[i+2 : i+4])
		if b == a+1 || b == a-1 {
			score += 15
		}
	}

	for i := 0; i <= n-4; i++ {
		a := number[i]
		b := number[i+1]
		c := number[i+2]
		d := number[i+3]
		if a == d && b == c {
			score += 20
		}
	}

	return score
}
