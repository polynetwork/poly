/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */

package utils

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	PARAMS_SPLIT          = ","
	PARAM_TYPE_SPLIT      = ":"
	PARAM_TYPE_ARRAY      = "array"
	PARAM_TYPE_BYTE_ARRAY = "bytearray"
	PARAM_TYPE_STRING     = "string"
	PARAM_TYPE_INTEGER    = "int"
	PARAM_TYPE_BOOLEAN    = "bool"
	PARAM_LEFT_BRACKET    = "["
	PARAM_RIGHT_BRACKET   = "]"
	PARAM_ESC_CHAR        = `/`
)

var (
	PARAM_TYPE_SPLIT_INC = string([]byte{0})
	PARAM_SPLIT_INC      = string([]byte{1})
)

//ParseParams return interface{} array of encode params item.
//A param item compose of type and value, type can be: bytearray, string, int, bool
//Param type and param value split with ":", such as int:10
//Param array can be express with "[]", such [int:10,string:foo], param array can be nested, such as [int:10,[int:12,bool:true]]
//A raw params example: string:foo,[int:0,[bool:true,string:bar],bool:false]
//Note that if string contain some special char like :,[,] and so one, please '/' char to escape. For example: string:did/:ed1e25c9dccae0c694ee892231407afa20b76008
func ParseParams(rawParamStr string) ([]interface{}, error) {
	rawParams, _, err := parseRawParamsString(rawParamStr)
	if err != nil {
		return nil, err
	}
	return parseRawParams(rawParams)
}

func parseRawParamsString(rawParamStr string) ([]interface{}, int, error) {
	if len(rawParamStr) == 0 {
		return nil, 0, nil
	}
	rawParamItems := make([]interface{}, 0)
	curRawParam := ""
	index := 0
	totalSize := len(rawParamStr)
	isEscMode := false
	for i := 0; i < totalSize; i++ {
		s := string(rawParamStr[i])
		if s == PARAM_ESC_CHAR {
			if isEscMode {
				isEscMode = false
				curRawParam += s
			} else {
				isEscMode = true
			}
			continue
		} else {
			if isEscMode {
				isEscMode = false
				curRawParam += s
				continue
			}
		}

		switch s {
		case PARAM_TYPE_SPLIT:
			curRawParam += PARAM_TYPE_SPLIT_INC
		case PARAMS_SPLIT:
			curRawParam = strings.TrimSpace(curRawParam)
			if len(curRawParam) > 0 {
				rawParamItems = append(rawParamItems, curRawParam)
				curRawParam = ""
			}
		case PARAM_LEFT_BRACKET:
			if index == totalSize-1 {
				return rawParamItems, 0, nil
			}
			//clear current param as invalid input
			curRawParam = ""
			items, size, err := parseRawParamsString(string(rawParamStr[i+1:]))
			if err != nil {
				return nil, 0, fmt.Errorf("parse params error:%s", err)
			}
			if len(items) > 0 {
				rawParamItems = append(rawParamItems, items)
			}
			i += size
		case PARAM_RIGHT_BRACKET:
			curRawParam = strings.TrimSpace(curRawParam)
			if len(curRawParam) > 0 {
				rawParamItems = append(rawParamItems, curRawParam)
			}
			return rawParamItems, i + 1, nil
		default:
			curRawParam += s
		}
	}
	curRawParam = strings.TrimSpace(curRawParam)
	if len(curRawParam) != 0 {
		rawParamItems = append(rawParamItems, curRawParam)
	}
	return rawParamItems, totalSize, nil
}

func parseRawParams(rawParams []interface{}) ([]interface{}, error) {
	if len(rawParams) == 0 {
		return nil, nil
	}
	params := make([]interface{}, 0)
	for _, rawParam := range rawParams {
		switch v := rawParam.(type) {
		case string:
			param, err := parseRawParam(v)
			if err != nil {
				return nil, err
			}
			params = append(params, param)
		case []interface{}:
			res, err := parseRawParams(v)
			if err != nil {
				return nil, err
			}
			params = append(params, res)
		default:
			return nil, fmt.Errorf("unknown param type:%s", reflect.TypeOf(rawParam))
		}
	}
	return params, nil
}

func parseRawParam(rawParam string) (interface{}, error) {
	rawParam = strings.TrimSpace(rawParam)
	//rawParam = strings.Trim(rawParam, PARAMS_SPLIT)
	if len(rawParam) == 0 {
		return nil, nil
	}
	ps := strings.Split(rawParam, PARAM_TYPE_SPLIT_INC)
	if len(ps) != 2 {
		return nil, fmt.Errorf("invalid param:%s", rawParam)
	}
	pType := strings.TrimSpace(ps[0])
	pValue := strings.TrimSpace(ps[1])
	return parseRawParamValue(pType, pValue)
}

func parseRawParamValue(pType string, pValue string) (interface{}, error) {
	switch strings.ToLower(pType) {
	case PARAM_TYPE_BYTE_ARRAY:
		value, err := hex.DecodeString(pValue)
		if err != nil {
			return nil, fmt.Errorf("parse byte array param:%s error:%s", pValue, err)
		}
		return value, nil
	case PARAM_TYPE_STRING:
		return pValue, nil
	case PARAM_TYPE_INTEGER:
		if pValue == "" {
			return nil, fmt.Errorf("invalid integer")
		}
		value, err := strconv.ParseInt(pValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse integer param:%s error:%s", pValue, err)
		}
		return value, nil
	case PARAM_TYPE_BOOLEAN:
		switch strings.ToLower(pValue) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("parse boolean param:%s failed", pValue)
		}
	default:
		return nil, fmt.Errorf("unspport param type:%s", pType)
	}
}
