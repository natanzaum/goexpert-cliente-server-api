package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const cotacaoAPIURL = "https://economia.awesomeapi.com.br/json/last/"

type Cotacao struct {
	Code        string `json:"code" gorm:"primaryKey"`
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
	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Cotacao{})

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		BuscaCotacaoHandler(w, r, db)
	})
	log.Printf("Servidor rodando em: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func BuscaCotacaoHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	moeda := r.URL.Query().Get("moeda")
	if moeda == "" {
		http.Error(w, "Parâmetro 'moeda' é obrigatório", http.StatusBadRequest)
		return
	}

	cotacao, err := BuscaCotacao(moeda)
	if err != nil {
		http.Error(w, "Erro ao buscar cotação", http.StatusInternalServerError)
		return
	}

	db.WithContext(ctx).Create(&cotacao)
	if ctx.Err() != nil {
		log.Printf("Erro ao salvar cotação no banco de dados: %v", ctx.Err())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cotacao)
}

func BuscaCotacao(moeda string) (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, cotacaoAPIURL+moeda, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("timeout excedido ao buscar cotação da moeda %s: %v", moeda, err)
		}
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]Cotacao
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	for _, cotacao := range result {
		return &cotacao, nil
	}

	return nil, fmt.Errorf("resposta da API está em formato inesperado")
}
