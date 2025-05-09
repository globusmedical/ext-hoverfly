package matchers

import (
	"bytes"
	"encoding/json"
)

var JsonPartial = "jsonpartial"

func JsonPartialMatch(match interface{}, toMatch string) bool {
	var expected interface{}
	var toMatchType interface{}
	matchString, ok := match.(string)
	if !ok {
		return false
	}

	d := json.NewDecoder(bytes.NewBuffer([]byte(toMatch)))
	d.UseNumber()
	err0 := d.Decode(&toMatchType)
	d = json.NewDecoder(bytes.NewBuffer([]byte(matchString)))
	d.UseNumber()
	err1 := d.Decode(&expected)
	if err0 != nil || err1 != nil {
		return false
	}

	actual, ok := toMatchType.(map[string]interface{})
	var allNodes []interface{}
	if ok {
		allNodes = getAllNodesFromMap(actual)
	} else {
		actual := toMatchType.([]interface{})
		allNodes = getAllNodesFromArray(actual)
	}

	for _, node := range allNodes {
		if expectedMap, ok := expected.(map[string]interface{}); ok {
			if mapContainsMap(expectedMap, node) {
				return true
			}
		} else if expectedArray, ok := expected.([]interface{}); ok {
			if arrayContainsArray(expectedArray, node) {
				return true
			}
		}
	}

	return false
}

func mapContainsMap(match, toMatch interface{}) bool {
	var expected, actual map[string]interface{}
	expected, ok0 := match.(map[string]interface{})
	actual, ok1 := toMatch.(map[string]interface{})
	if !ok0 || !ok1 {
		return false
	}
	for key, val := range expected {
		if innerMap, ok := val.(map[string]interface{}); ok {
			if !mapContainsMap(innerMap, actual[key]) {
				return false
			}
		} else if innerArr, ok := val.([]interface{}); ok {
			if !arrayContainsArray(innerArr, actual[key]) {
				return false
			}
		} else {
			actualValue, exist := actual[key]
			if !exist || actualValue != val {
				return false
			}
		}
	}
	return true
}

func arrayContainsArray(match, toMatch interface{}) bool {

	var expected, actual []interface{}
	expected, ok0 := match.([]interface{})
	actual, ok1 := toMatch.([]interface{})
	if !ok0 || !ok1 {
		return false
	}

	for _, cur := range expected {
		if innerMap, ok := cur.(map[string]interface{}); ok {
			result := arrContainsMap(actual, innerMap)
			if !result {
				return false
			}
		} else if innerArr, ok := cur.([]interface{}); ok {
			if !arrayContainsArray(innerArr, actual) {
				return false
			}
		} else {
			if !arrContainsObj(actual, cur) {
				return false
			}
		}
	}
	return true
}

func arrContainsObj(arr []interface{}, obj interface{}) bool {
	for _, val := range arr {
		if val == obj {
			return true
		}
	}
	return false
}

func arrContainsMap(arr []interface{}, mp map[string]interface{}) bool {
	for _, val := range arr {
		if innerMap, ok := val.(map[string]interface{}); ok {
			if mapContainsMap(mp, innerMap) {
				return true
			}
		} else if innerArr, ok := val.([]interface{}); ok {
			if arrContainsMap(innerArr, mp) {
				return true
			}
		}
	}
	return false
}

func getAllNodesFromMap(current map[string]interface{}) []interface{} {
	var allNodes = make([]interface{}, 0)
	allNodes = append(allNodes, current)
	for _, val := range current {
		if innerMap, ok := val.(map[string]interface{}); ok {
			allNodes = append(allNodes, getAllNodesFromMap(innerMap)...)
		} else if innerArray, ok := val.([]interface{}); ok {
			allNodes = append(allNodes, getAllNodesFromArray(innerArray)...)
		}
	}
	return allNodes
}

func getAllNodesFromArray(current []interface{}) []interface{} {
	var allNodes = make([]interface{}, 0)
	allNodes = append(allNodes, current)
	for _, val := range current {
		if innerMap, ok := val.(map[string]interface{}); ok {
			allNodes = append(allNodes, getAllNodesFromMap(innerMap)...)
		} else if innerArray, ok := val.([]interface{}); ok {
			allNodes = append(allNodes, getAllNodesFromArray(innerArray)...)
		}
	}
	return allNodes
}
