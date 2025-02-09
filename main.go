package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Address struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Cidade     string `json:"localidade"`
	Estado     string `json:"uf"`
	Source     string `json:"source"`
}

type BrasilApiAddress struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCepAddress struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Estado      string `json:"estado"`
	Regiao      string `json:"regiao"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

func fetchAPI(url string, source string, resultChan chan<- Address, errChan chan<- error) {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		errChan <- err
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("erro ao acessar %s: status %d", source, resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errChan <- err
		return
	}

	var resultAddress Address

	switch source {
	case "ViaCEP":
		var address ViaCepAddress
		if err := json.Unmarshal(body, &address); err != nil {
			errChan <- err
			return
		}
		resultAddress = Address{
			Cep:        address.Cep,
			Logradouro: address.Logradouro,
			Bairro:     address.Bairro,
			Cidade:     address.Localidade,
			Estado:     address.Uf,
			Source:     source,
		}
	case "BrasilAPI":
		var address BrasilApiAddress
		if err := json.Unmarshal(body, &address); err != nil {
			errChan <- err
			return
		}
		resultAddress = Address{
			Cep:        address.Cep,
			Logradouro: address.Street,
			Bairro:     address.Neighborhood,
			Cidade:     address.City,
			Estado:     address.State,
			Source:     source,
		}
	}

	resultChan <- resultAddress

	fmt.Printf("%s respondeu em %v\n", source, time.Since(start))
}

func main() {
	cep := "68904360"
	brasilAPI := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	viaCEP := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	resultChan := make(chan Address, 1)
	errChan := make(chan error, 2)
	go fetchAPI(viaCEP, "ViaCEP", resultChan, errChan)
	go fetchAPI(brasilAPI, "BrasilAPI", resultChan, errChan)

	select {
	case result := <-resultChan:
		fmt.Printf("Resposta mais rápida:\nAPI: %s\nCEP: %s\nLogradouro: %s\nBairro: %s\nCidade: %s\nEstado: %s\n",
			result.Source, result.Cep, result.Logradouro, result.Bairro, result.Cidade, result.Estado)
	case err := <-errChan:
		fmt.Println("Erro:", err)
	case <-time.After(1 * time.Second):
		fmt.Println("Erro: Timeout ao buscar o CEP")
	}
}
