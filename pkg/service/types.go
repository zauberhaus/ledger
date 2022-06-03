package service

type Holder struct {
	Name     string
	Accounts []*Account
}

type Account struct {
	Account string
	Asset   string
}

type Asset struct {
	Symbol string
	Name   string
}

type Assets []Asset

func (a Assets) Len() int           { return len(a) }
func (a Assets) Less(i, j int) bool { return a[i].Symbol < a[j].Symbol }
func (a Assets) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type Status struct {
	ID   int
	Name string
}

type Statuses []Status

func (s Statuses) Len() int           { return len(s) }
func (s Statuses) Less(i, j int) bool { return s[i].ID < s[j].ID }
func (s Statuses) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
