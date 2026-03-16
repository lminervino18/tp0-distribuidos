package common

// Bet representa una apuesta de una agencia de quiniela

type Bet struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    string
}

func NewBet(agency, firstName, lastName, document, birthdate, number string) *Bet {
	return &Bet{
		Agency:    agency,
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}
}
