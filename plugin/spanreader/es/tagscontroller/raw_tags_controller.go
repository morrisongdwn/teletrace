package tagscontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"oss-tracing/pkg/model/tagsquery/v1"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.uber.org/zap"
)

type rawTagsController struct {
	client *elasticsearch.Client
	idx    string
}

func NewTagsController(logger *zap.Logger, client *elasticsearch.Client, idx string) (TagsController, error) {
	return &rawTagsController{
		client: client,
		idx:    idx,
	}, nil
}

func (r *rawTagsController) prefixTag(tag string) string {
	return fmt.Sprint("span.attributes.", tag)
}

// Get available tags.
// Use elastic's GetFieldMapping api
func (r *rawTagsController) GetAvailableTags(
	ctx context.Context,
	request tagsquery.GetAvailableTagsRequest,
) (tagsquery.GetAvailableTagsResponse, error) {
	result := tagsquery.GetAvailableTagsResponse{
		Tags: make([]tagsquery.TagInfo, 0),
	}

	res, err := r.client.Indices.GetFieldMapping(
		[]string{r.prefixTag("*")},
		r.client.Indices.GetFieldMapping.WithContext(ctx),
		r.client.Indices.GetFieldMapping.WithIndex(r.idx),
	)

	if err != nil {
		return result, fmt.Errorf("failed to get field mapping: %v", err)
	}

	defer res.Body.Close()
	if err := SummarizeResponseError(res); err != nil {
		return result, err
	}

	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return result, fmt.Errorf("failed to decode body: %v", err)
	}

	// { "lupa-index": { "mappings": { ... }}}
	body = body[r.idx].(map[string]any)["mappings"].(map[string]any)

	for _, v := range body {
		fieldMapping := v.(map[string]any)

		mappingData := fieldMapping["mapping"].(map[string]any)
		if len(mappingData) > 1 {
			return result, fmt.Errorf(
				"unknown scenario - mapping data contains more than one entry: %v", mappingData)
		}

		for _, valueData := range mappingData {
			result.Tags = append(result.Tags, tagsquery.TagInfo{
				Name: fieldMapping["full_name"].(string),
				Type: valueData.(map[string]any)["type"].(string),
			})
		}
	}

	return result, nil
}

func (r *rawTagsController) GetTagsValues(
	ctx context.Context,
	request tagsquery.TagValuesRequest,
	tags []string,
) (map[string]*tagsquery.TagValuesResponse, error) {

	body, err := r.performGetTagsValuesRequest(ctx, request, tags)

	if err != nil {
		return map[string]*tagsquery.TagValuesResponse{}, err
	}

	return r.parseGetTagsValuesResponseBody(body)
}

// Perform search and return the response body
func (r *rawTagsController) performGetTagsValuesRequest(
	ctx context.Context,
	request tagsquery.TagValuesRequest,
	tags []string,
) (map[string]any, error) {
	aggs := make(map[string]any, len(tags))
	for _, field := range tags {

		aggregationKey := field

		aggs[aggregationKey] = map[string]any{
			"terms": map[string]any{
				"field": fmt.Sprintf("%s.keyword", field),
			},
		}
	}

	query := map[string]any{
		"bool": map[string]any{
			"must": []map[string]any{
				{
					"range": map[string]any{
						"span.startTimeUnixNano": map[string]any{
							"gte": request.Timeframe.StartTime,
						},
					},
				},
				{
					"range": map[string]any{
						"span.endTimeUnixNano": map[string]any{
							"lte": request.Timeframe.EndTime,
						},
					},
				},
			},
		},
	}

	requestBody := map[string]any{
		"size":  0,
		"aggs":  aggs,
		"query": query,
	}

	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(requestBody); err != nil {
		return nil, fmt.Errorf("failed to encode body; %v: %v", requestBody, err)
	}

	searchOptions := []func(*esapi.SearchRequest){
		r.client.Search.WithIndex(r.idx),
		r.client.Search.WithBody(buffer),
	}

	res, err := r.client.Search(searchOptions...)

	if err != nil {
		return nil, fmt.Errorf("failed to perform search: %s", err)
	}

	defer res.Body.Close()
	if err := SummarizeResponseError(res); err != nil {
		return nil, err
	}

	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed parsing the response body: %s", err)
	}

	return body, err
}

func (r *rawTagsController) parseGetTagsValuesResponseBody(
	body map[string]any,
) (map[string]*tagsquery.TagValuesResponse, error) {

	// To get an idea of how the response looks like, check the unit test at raw_tags_controller_test.go

	result := map[string]*tagsquery.TagValuesResponse{}
	aggregations := body["aggregations"].(map[string]any)
	tagValueInfos := make(map[string]map[any]tagsquery.TagValueInfo)

	// the aggregation key is the tag's name because that's how we defined the query.
	// traverse the returned aggregations, bucket by bucket and update the value counts
	for tag, v := range aggregations {
		aggregation := v.(map[string]any)

		if _, found := tagValueInfos[tag]; !found {
			tagValueInfos[tag] = make(map[any]tagsquery.TagValueInfo)
		}

		for _, v := range aggregation["buckets"].([]any) {
			bucket := v.(map[string]any)
			value := bucket["key"]
			count := int(bucket["doc_count"].(float64))

			if info, found := tagValueInfos[tag][value]; !found {
				tagValueInfos[tag][value] = tagsquery.TagValueInfo{
					Value: value,
					Count: count,
				}
			} else {
				info.Count += count
				tagValueInfos[tag][value] = info
			}
		}
	}

	// populate the result
	for tag, valueInfoMap := range tagValueInfos {
		var currentTagValues []tagsquery.TagValueInfo
		for value, info := range valueInfoMap {
			currentTagValues = append(currentTagValues, tagsquery.TagValueInfo{
				Value: value,
				Count: info.Count,
			})
		}
		if currentTagValues != nil {
			result[tag] = &tagsquery.TagValuesResponse{
				Values: currentTagValues,
			}
		}
	}

	return result, nil
}