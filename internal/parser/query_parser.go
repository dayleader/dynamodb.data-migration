package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"dynamodb.data-migration/internal/domain"
)

type parser struct {
}

// NewQueryParser - constructs a new query parser.
func NewQueryParser() domain.QueryParser {
	return &parser{}
}

func (p *parser) ParseContent(content []byte) ([]*domain.DynamoDBQuery, error) {
	if len(content) == 0 {
		return nil, errors.New("Cannot parse empty query content")
	}
	var mapSlice []map[string]interface{}
	if err := json.Unmarshal(content, &mapSlice); err != nil {
		return nil, err
	}
	result := make([]*domain.DynamoDBQuery, len(mapSlice))
	for i, m := range mapSlice {
		tableName, ok := convertToString(m[domain.JSONFieldTableName])
		if !ok {
			return nil, errors.New("Cannot parse table name")
		}
		schema := []*domain.DynamoDBSchema{}
		if _, ok := m[domain.JSONFieldSchema]; ok {
			err := fillStruct(&schema, m[domain.JSONFieldSchema])
			if err != nil {
				return nil, err
			}
		}
		data := []map[string]interface{}{}
		if _, ok := m[domain.JSONFieldData]; ok {
			data, ok = convertToSliceMap(m[domain.JSONFieldData])
			if !ok {
				return nil, fmt.Errorf("Cannot parse data for %s", tableName)
			}
		}
		result[i] = &domain.DynamoDBQuery{
			TableName: tableName,
			Schema:    schema,
			Data:      data,
		}
	}
	return result, nil
}

func convertToSliceMap(val interface{}) ([]map[string]interface{}, bool) {
	if val == nil {
		return []map[string]interface{}{}, false
	}
	if singleVal, ok := convertToMap(val); ok {
		return []map[string]interface{}{singleVal}, ok
	}
	if mapArray, ok := val.([]map[string]interface{}); ok {
		return mapArray, ok
	}
	slice, ok := val.([]interface{})
	if ok {
		mapSlice := make([]map[string]interface{}, len(slice))
		for i, iVal := range slice {
			mapValue, ok := convertToMap(iVal)
			if !ok {
				return []map[string]interface{}{}, false
			}
			mapSlice[i] = mapValue
		}
		return mapSlice, ok
	}
	return []map[string]interface{}{}, false
}

func convertToMap(val interface{}) (map[string]interface{}, bool) {
	mapVal, ok := val.(map[string]interface{})
	if ok {
		return mapVal, ok
	}
	iMapVal, ok := val.(map[interface{}]interface{})
	if ok {
		m := make(map[string]interface{}, len(iMapVal))
		for i, sVal := range iMapVal {
			iVal, ok := convertToString(i)
			if !ok {
				return nil, false
			}
			m[iVal] = sVal
		}
		return m, true
	}
	return map[string]interface{}{}, false
}

func convertToString(v interface{}) (string, bool) {
	stringValue, ok := v.(string)
	return stringValue, ok
}

func fillStruct(any interface{}, m interface{}) error {
	if err := checkIsPointer(any); err != nil {
		return err
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, any)
	return err
}

func checkIsPointer(any interface{}) error {
	if reflect.ValueOf(any).Kind() != reflect.Ptr {
		return fmt.Errorf("You passed something that was not a pointer: %s", any)
	}
	return nil
}
