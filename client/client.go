package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Code        string `json:"code"`
	Codein      string `json:"codein"`
	Name        string `json:"name"`
	High        string `json:"high"`
	Low         string `json:"low"`
	VarBid      string `json:"varBid"`
	PctChange   string `json:"pctChange"`
	Bid         string `json:"bid"`
	Ask         string `json:"ask"`
	Timestamp   string `json:"timestamp"`
	Create_date string `json:"create_date"`
}

func main() {
	//faz uma requisição http para o servidor cm contexto de timeout
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	//cria o cliente que faz o request e usa o contexto com timeout
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao?moeda=USD-BRL", nil)
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			//se o contexto tiver expirado, significa que o timeout foi atingido
			println("Timeout atingido ao buscar cotação")
		} else {
			panic(err)
		}
		return
	}
	defer resp.Body.Close()

	//salva somente o campo bid em um arquivo de texto chamado cotacao.txt
	var cotacao Cotacao
	err = json.NewDecoder(resp.Body).Decode(&cotacao)
	if err != nil {
		panic(err)
	}

	//salva o campo bid em um arquivo de texto chamado cotacao.txt
	err = os.WriteFile("cotacao.txt", []byte("Dolar: "+cotacao.Bid), 0644)
	if err != nil {
		panic(err)
	}
}
