package models

import "github.com/google/uuid"

type Client struct {
	ID   uuid.UUID
	Name string
}

func NewClient(name string) *Client {
	return &Client{ID: uuid.New(), Name: name}
}

type Cluster struct {
	ID       uuid.UUID
	Client   *Client
	Provider string
}
