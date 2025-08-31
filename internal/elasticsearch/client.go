package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/config"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
	"github.com/terminator791/Simple-elasticsearch-GO/pkg/logger"
)

const (
	ProductsIndex = "products"
)

// Client wraps the Elasticsearch client
type Client struct {
	es *elasticsearch.Client
}

// NewClient creates a new Elasticsearch client
func NewClient(cfg *config.ElasticsearchConfig) (*Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.Host},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch returned error: %s", res.Status())
	}

	logger.InfoLogger.Println("Successfully connected to Elasticsearch")

	client := &Client{es: es}

	// Create index with proper mapping
	if err := client.createProductsIndex(); err != nil {
		return nil, fmt.Errorf("failed to create products index: %w", err)
	}

	return client, nil
}

// createProductsIndex creates the products index with explicit mapping
func (c *Client) createProductsIndex() error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"name": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"description": map[string]interface{}{
					"type": "text",
				},
				"brand": map[string]interface{}{
					"type": "keyword",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"price": map[string]interface{}{
					"type": "float",
				},
				"stock_quantity": map[string]interface{}{
					"type": "integer",
				},
				"created_at": map[string]interface{}{
					"type":   "date",
					"format": "strict_date_optional_time||epoch_millis",
				},
				"updated_at": map[string]interface{}{
					"type":   "date",
					"format": "strict_date_optional_time||epoch_millis",
				},
			},
		},
	}

	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: ProductsIndex,
		Body:  bytes.NewReader(mappingJSON),
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	// Index already exists is not an error
	if res.IsError() && !strings.Contains(res.Status(), "400") {
		return fmt.Errorf("failed to create index: %s", res.Status())
	}

	logger.InfoLogger.Printf("Products index created/verified successfully")
	return nil
}

// IndexProduct indexes a product in Elasticsearch
func (c *Client) IndexProduct(product *models.Product) error {
	productJSON, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      ProductsIndex,
		DocumentID: product.ID.String(),
		Body:       bytes.NewReader(productJSON),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return fmt.Errorf("failed to index product: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index product: %s", res.Status())
	}

	return nil
}

// DeleteProduct deletes a product from Elasticsearch
func (c *Client) DeleteProduct(id string) error {
	req := esapi.DeleteRequest{
		Index:      ProductsIndex,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && !strings.Contains(res.Status(), "404") {
		return fmt.Errorf("failed to delete product: %s", res.Status())
	}

	return nil
}

// SearchProducts performs complex search operations
func (c *Client) SearchProducts(searchReq *models.SearchRequest) (*models.SearchResponse, time.Duration, error) {
	start := time.Now()

	// Build the search query
	query := c.buildSearchQuery(searchReq)

	// Add aggregations
	aggs := map[string]interface{}{
		"brands": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "brand",
				"size":  20,
			},
		},
		"categories": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "category",
				"size":  20,
			},
		},
	}

	// Build sort
	sort := c.buildSort(searchReq)

	searchBody := map[string]interface{}{
		"query": query,
		"aggs":  aggs,
		"from":  (searchReq.Page - 1) * searchReq.Size,
		"size":  searchReq.Size,
		"sort":  sort,
	}

	searchJSON, err := json.Marshal(searchBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal search query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{ProductsIndex},
		Body:  bytes.NewReader(searchJSON),
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search products: %w", err)
	}
	defer res.Body.Close()

	duration := time.Since(start)

	if res.IsError() {
		return nil, duration, fmt.Errorf("search request failed: %s", res.Status())
	}

	var searchResponse ElasticsearchResponse
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, duration, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Convert to our response format
	products := make([]models.Product, len(searchResponse.Hits.Hits))
	for i, hit := range searchResponse.Hits.Hits {
		products[i] = hit.Source
	}

	aggregations := make(map[string]models.Aggregation)
	if searchResponse.Aggregations.Brands != nil {
		buckets := make([]models.Bucket, len(searchResponse.Aggregations.Brands.Buckets))
		for i, bucket := range searchResponse.Aggregations.Brands.Buckets {
			buckets[i] = models.Bucket{
				Key:   bucket.Key,
				Count: bucket.DocCount,
			}
		}
		aggregations["brands"] = models.Aggregation{Buckets: buckets}
	}

	if searchResponse.Aggregations.Categories != nil {
		buckets := make([]models.Bucket, len(searchResponse.Aggregations.Categories.Buckets))
		for i, bucket := range searchResponse.Aggregations.Categories.Buckets {
			buckets[i] = models.Bucket{
				Key:   bucket.Key,
				Count: bucket.DocCount,
			}
		}
		aggregations["categories"] = models.Aggregation{Buckets: buckets}
	}

	response := &models.SearchResponse{
		Products:     products,
		Total:        searchResponse.Hits.Total.Value,
		Page:         searchReq.Page,
		Size:         searchReq.Size,
		Aggregations: aggregations,
	}

	return response, duration, nil
}

// buildSearchQuery builds the Elasticsearch query based on search parameters
func (c *Client) buildSearchQuery(searchReq *models.SearchRequest) map[string]interface{} {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{}

	// Full-text search on name and description
	if searchReq.Query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  searchReq.Query,
				"fields": []string{"name^2", "description"},
				"type":   "best_fields",
			},
		})
	}

	// Brand filter
	if len(searchReq.Brand) > 0 {
		filter = append(filter, map[string]interface{}{
			"terms": map[string]interface{}{
				"brand": searchReq.Brand,
			},
		})
	}

	// Category filter
	if len(searchReq.Category) > 0 {
		filter = append(filter, map[string]interface{}{
			"terms": map[string]interface{}{
				"category": searchReq.Category,
			},
		})
	}

	// Price range filter
	if searchReq.MinPrice != nil || searchReq.MaxPrice != nil {
		rangeQuery := map[string]interface{}{}
		if searchReq.MinPrice != nil {
			rangeQuery["gte"] = *searchReq.MinPrice
		}
		if searchReq.MaxPrice != nil {
			rangeQuery["lte"] = *searchReq.MaxPrice
		}
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{
				"price": rangeQuery,
			},
		})
	}

	// Build the final query
	if len(must) == 0 && len(filter) == 0 {
		return map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	boolQuery := map[string]interface{}{}
	if len(must) > 0 {
		boolQuery["must"] = must
	}
	if len(filter) > 0 {
		boolQuery["filter"] = filter
	}

	return map[string]interface{}{
		"bool": boolQuery,
	}
}

// buildSort builds the sort clause for Elasticsearch
func (c *Client) buildSort(searchReq *models.SearchRequest) []map[string]interface{} {
	if searchReq.SortBy == "" {
		searchReq.SortBy = "created_at"
	}
	if searchReq.SortOrder == "" {
		searchReq.SortOrder = "desc"
	}

	sortField := searchReq.SortBy
	if sortField == "name" {
		sortField = "name.keyword"
	}

	return []map[string]interface{}{
		{
			sortField: map[string]interface{}{
				"order": searchReq.SortOrder,
			},
		},
	}
}

// BulkIndex performs bulk indexing of products
func (c *Client) BulkIndex(products []models.Product) error {
	if len(products) == 0 {
		return nil
	}

	var buffer bytes.Buffer

	for _, product := range products {
		// Index action
		indexAction := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": ProductsIndex,
				"_id":    product.ID.String(),
			},
		}

		actionJSON, err := json.Marshal(indexAction)
		if err != nil {
			return fmt.Errorf("failed to marshal index action: %w", err)
		}

		productJSON, err := json.Marshal(product)
		if err != nil {
			return fmt.Errorf("failed to marshal product: %w", err)
		}

		buffer.Write(actionJSON)
		buffer.WriteByte('\n')
		buffer.Write(productJSON)
		buffer.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body:    &buffer,
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return fmt.Errorf("failed to perform bulk index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index failed: %s", res.Status())
	}

	logger.InfoLogger.Printf("Successfully bulk indexed %d products", len(products))
	return nil
}

// ElasticsearchResponse represents the response from Elasticsearch
type ElasticsearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source models.Product `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		Brands *struct {
			Buckets []struct {
				Key      string `json:"key"`
				DocCount int64  `json:"doc_count"`
			} `json:"buckets"`
		} `json:"brands"`
		Categories *struct {
			Buckets []struct {
				Key      string `json:"key"`
				DocCount int64  `json:"doc_count"`
			} `json:"buckets"`
		} `json:"categories"`
	} `json:"aggregations"`
}