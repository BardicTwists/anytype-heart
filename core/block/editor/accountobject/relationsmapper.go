package accountobject

import (
	"github.com/gogo/protobuf/types"
	"github.com/valyala/fastjson"

	"github.com/anyproto/anytype-heart/util/pbtypes"
)

type KeyType int

const (
	KeyTypeString KeyType = iota
	KeyTypeInt64
)

type relationsMapper struct {
	keys map[string]KeyType
}

func newRelationsMapper(keys map[string]KeyType) *relationsMapper {
	return &relationsMapper{
		keys: keys,
	}
}

func (r *relationsMapper) GetRelationKey(key string, val *fastjson.Value) (*types.Value, bool) {
	kt, ok := r.keys[key]
	if !ok {
		return nil, false
	}
	switch kt {
	case KeyTypeString:
		val := val.GetStringBytes(key)
		if val == nil {
			return nil, false
		}
		return pbtypes.String(string(val)), true
	case KeyTypeInt64:
		val := val.GetInt64(key)
		if val == 0 {
			return nil, false
		}
		return pbtypes.Int64(val), true
	case KeyTypeBool:
		return pbtypes.Bool(val.GetBool(key)), true
	}
	return nil, false
}

func (r *relationsMapper) GetStoreKey(key string, val *types.Value) (res any, ok bool) {
	kt, ok := r.keys[key]
	if !ok {
		return nil, false
	}
	switch kt {
	case KeyTypeString:
		res = val.GetStringValue()
		if res == "" {
			return nil, false
		}
	case KeyTypeInt64:
		res = int64(val.GetNumberValue())
		if res == 0 {
			return nil, false
		}
	case KeyTypeBool:
		res = val.GetBoolValue()
	}
	return res, true
}
