package commands

import "encoding/json"

type DomainsCommand struct{}

func (command *DomainsCommand) Execute([]string) error {
	logger := Veritas.logger.Session("domains")
	encoder := json.NewEncoder(Veritas.output)

	domains, err := Veritas.bbsClient.Domains(logger)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		encoder.Encode(domain)
	}

	return nil
}
