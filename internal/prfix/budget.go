package prfix

import "fmt"

type Budget struct {
	Limit int
	Used  int
}

func (budget *Budget) Take() error {
	if budget == nil {
		return nil
	}
	if budget.Limit < 1 || budget.Used >= budget.Limit {
		return fmt.Errorf("GitHub request budget of %d is exhausted", budget.Limit)
	}
	budget.Used++
	return nil
}
