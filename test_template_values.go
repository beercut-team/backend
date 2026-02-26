package main

import (
	"fmt"
	"github.com/beercut-team/backend-boilerplate/internal/domain"
)

func main() {
	templates := domain.GetChecklistTemplates(domain.OperationPhacoemulsification)
	
	fmt.Println("Template definitions for PHACOEMULSIFICATION:")
	for i, t := range templates {
		fmt.Printf("%d. %s - IsRequired: %v\n", i+1, t.Name, t.IsRequired)
	}
}
