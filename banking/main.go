package main

type SavingsAccount struct {
	balance int
}

func (sa *SavingsAccount) MonthlyInterest() int {
	yearlyInterest := float64(sa.balance) * 0.05
	return int(yearlyInterest / 12)
}

func (sa *SavingsAccount) Transfer(receiver Account, amount int) string {
	if amount <= 0 {
		return "Amount cannot be negative"
	}
	switch receiver.(type) {
	case *SavingsAccount, *CheckingAccount, *InvestmentAccount:
	default:
		return "Invalid receiver account"
	}
	if sa.balance < amount {
		return "Account balance is not enough"
	}
	sa.balance = sa.balance - amount
	receiver.Deposit(amount)
	return "Success"

}

func (sa *SavingsAccount) Deposit(amount int) string {
	if amount <= 0 {
		return "Amount cannot be negative"
	}
	sa.balance += amount
	return "Success"
}

func (sa *SavingsAccount) Withdraw(amount int) string {
	if amount <= 0 {
		return "Amount cannot be negative"
	}
	if sa.balance < amount {
		return "Account balance is not enough"
	}
	sa.balance -= amount
	return "Success"
}

func (sa *SavingsAccount) CheckBalance() int {
	return int(sa.balance)
}

// ---

type CheckingAccount struct {
	balance int
}

func (ca *CheckingAccount) MonthlyInterest() int {
	yearlyInterest := float64(ca.balance) * 0.05
	return int(yearlyInterest / 12)
}

func (ca *CheckingAccount) Transfer(receiver Account, amount int) string {
	if amount <= 0 {
		return "Amount cannot be negative"
	}
	switch receiver.(type) {
	case *SavingsAccount, *CheckingAccount, *InvestmentAccount:
	default:
		return "Invalid receiver account"
	}
	if ca.balance < amount {
		return "Account balance is not enough"
	}
	ca.balance = ca.balance - amount
	receiver.Deposit(amount)
	return "Success"

}
func (ca *CheckingAccount) Deposit(amount int) string {
	if amount <= 0 {
		return "Amount cannot be negative"
	}
	ca.balance += amount
	return "Success"
}
func (ca *CheckingAccount) Withdraw(amount int) string {
	if amount <= 0 {
		return "Amount cannot be negative"
	}
	if ca.balance < amount {
		return "Account balance is not enough"
	}
	ca.balance -= amount
	return "Success"
}

func (ca *CheckingAccount) CheckBalance() int {
	return int(ca.balance)
}

// ---

type InvestmentAccount struct {
	balance int
}

func (ia *InvestmentAccount) MonthlyInterest() int {
	yearlyInterest := float64(ia.balance) * 0.05
	return int(yearlyInterest / 12)
}

func (ia *InvestmentAccount) Transfer(receiver Account, amount int) string {
	if amount <= 0 {
		return "Amount cannot be negative"
	}
	switch receiver.(type) {
	case *SavingsAccount, *CheckingAccount, *InvestmentAccount:
	default:
		return "Invalid receiver account"
	}
	if ia.balance < amount {
		return "Account balance is not enough"
	}
	ia.balance = ia.balance - amount
	receiver.Deposit(amount)
	return "Success"

}

func (ia *InvestmentAccount) Deposit(amount int) string {
	if amount <= 0 {
		return "Amount cannot be negative"
	}
	ia.balance += amount
	return "Success"
}

func (ia *InvestmentAccount) Withdraw(amount int) string {
	if amount <= 0 {
		return "Amount cannot be negative"
	}
	if ia.balance < amount {
		return "Account balance is not enough"
	}
	ia.balance -= amount
	return "Success"
}

func (ia *InvestmentAccount) CheckBalance() int {
	return int(ia.balance)
}

// ---

func NewSavingsAccount() *SavingsAccount {
	return &SavingsAccount{
		balance: 0,
	}
}

func NewCheckingAccount() *CheckingAccount {
	return &CheckingAccount{
		balance: 0,
	}
}

func NewInvestmentAccount() *InvestmentAccount {
	return &InvestmentAccount{
		balance: 0,
	}
}
