package commands

import (
	"encoding/json"
	"errors"
)

type DomainsCommand struct{}

func (command *DomainsCommand) Execute([]string) error {
	logger := Veritas.logger.Session("domains")
	encoder := json.NewEncoder(Veritas.output)
	if Veritas.bbsClient == nil {
		return errors.New("error: the required flag `--bbsURL' was not specified")
	}

	domains, err := Veritas.bbsClient.Domains(logger)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		encoder.Encode(domain)
	}

	return nil
}
