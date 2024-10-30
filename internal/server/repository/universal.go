package repository

import "github.com/npavlov/go-metrics-service/internal/server/storage"

type Universal struct {
	Repo    Repository
	Storage storage.InMemory
}

func NewUniversal(repo Repository, storage storage.InMemory) *Universal {
	return &Universal{
		Repo:    repo,
		Storage: storage,
	}
}
