package test

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestDownloadLimit(t *testing.T) {
	client := &http.Client{}
	for i := 1; i <= 110; i++ {
		resp, err := client.Get("http://localhost:8003/api/white")
		if err != nil {
			t.Fatal(err)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Printf("Request %d: status=%d, size=%d bytes\n",
			i, resp.StatusCode, len(body))

		if resp.StatusCode == 429 {
			fmt.Printf("Лимит достигнут после %d запросов\n", i)
			break
		}
	}
}
