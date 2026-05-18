package utils

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

func RemoveByIndex(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}

func KeysSortedByValueDesc(m map[string]int) []string {
	type kv struct {
		Key   string
		Value int
	}
	ss := make([]kv, 0, len(m))
	for k, v := range m {
		ss = append(ss, kv{Key: k, Value: v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	keys := make([]string, 0, len(ss))
	for _, item := range ss {
		keys = append(keys, item.Key)
	}

	return keys
}

func ExtractBearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format, expected 'Bearer <token>'")
	}

	token := parts[1]
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}
