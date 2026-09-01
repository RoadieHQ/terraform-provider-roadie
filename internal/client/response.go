package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type DataWrapper[T any] struct {
	Data   T   `json:"data"`
	Total  int `json:"total,omitempty"`
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
}

type ItemsWrapper[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total,omitempty"`
}

func GetWrapped[T any](c *RoadieClient, ctx context.Context, path string) (*T, error) {
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var wrapper DataWrapper[T]
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &wrapper.Data, nil
}

func GetBare[T any](c *RoadieClient, ctx context.Context, path string) (*T, error) {
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func CreateWrapped[T any](c *RoadieClient, ctx context.Context, path string, input any) (*T, error) {
	body, err := c.Post(ctx, path, input)
	if err != nil {
		return nil, err
	}
	var wrapper DataWrapper[T]
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &wrapper.Data, nil
}

func CreateBare[T any](c *RoadieClient, ctx context.Context, path string, input any) (*T, error) {
	body, err := c.Post(ctx, path, input)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func UpdateWrapped[T any](c *RoadieClient, ctx context.Context, path string, input any) (*T, error) {
	body, err := c.Put(ctx, path, input)
	if err != nil {
		return nil, err
	}
	var wrapper DataWrapper[T]
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &wrapper.Data, nil
}

func UpdateBare[T any](c *RoadieClient, ctx context.Context, path string, input any) (*T, error) {
	body, err := c.Put(ctx, path, input)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func PatchBare[T any](c *RoadieClient, ctx context.Context, path string, input any) (*T, error) {
	body, err := c.Patch(ctx, path, input)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}
