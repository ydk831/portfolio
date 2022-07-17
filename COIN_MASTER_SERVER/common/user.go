package common

import (
	"time"
)

// 재화 타입 선언
const (
	GoodsTypeGold = iota + 0
	GoodsTypeShield
	GoodsTypeSpin
	GoodsTypeCoin
	GoodsTypeMax
)

const (
	BuildingEffectNone = iota + 0
	BuildingEffectGold
	BuildingEffectSpin
	BuildingEffectMaxSpin
	BuildingEffectMultiple
)

func IsValidType(v interface{}, t string) bool {
	switch v.(type) {
	case int8:
		return "int8" == t
	case int16:
		return "int16" == t
	case int:
		return "int" == t
	case int64:
		return "int64" == t
	case float32:
		return "float32" == t
	case float64:
		return "float64" == t
	case string:
		return "string" == t
	default:
		return false
	}
}

// CalcRechargeSP : SP 회복
func CalcRechargeSP(curSpin int16, maxSpin int16, now time.Time, rechargeDate time.Time) (int16, time.Time) {
	rechargeSec := int64(600000000000)
	rechargeSP := int16(now.Sub(rechargeDate) / time.Duration(rechargeSec)) // 10분에 1씩 회복

	if rechargeSP > 0 {
		if curSpin+rechargeSP >= maxSpin {
			curSpin = maxSpin
			rechargeDate = now
		} else {
			curSpin += rechargeSP
			rechargeDate = rechargeDate.Add(time.Duration(int64(rechargeSP) * rechargeSec))
		}
	}

	return curSpin, rechargeDate
}

func AddGoodsFromAllType(goodsType int, addValue int, curGold *int64, curSpin *int16, curShield *int8) {
	switch goodsType {
	case GoodsTypeGold:
		*curGold = AddGoods(goodsType, int64(addValue), *curGold).(int64)
	case GoodsTypeSpin:
		*curSpin = AddGoods(goodsType, int16(addValue), *curSpin).(int16)
	case GoodsTypeShield:
		*curShield = AddGoods(goodsType, int8(addValue), *curShield).(int8)
	default:
		panic("not define goods type!")
	}
}

// AddResource : 재화 획득
func AddGoods(goodsType int, addValue interface{}, curValue interface{}) interface{} {
	maxValue := getMaxGoodsValue(goodsType)

	switch curValue.(type) {
	case int64:
		if curValue.(int64)+addValue.(int64) < maxValue.(int64) {
			return curValue.(int64) + addValue.(int64)
		}
	case int32:
		if curValue.(int32)+addValue.(int32) < maxValue.(int32) {
			return curValue.(int32) + addValue.(int32)
		}
	case int16:
		if curValue.(int16)+addValue.(int16) < maxValue.(int16) {
			return curValue.(int16) + addValue.(int16)
		}
	case int8:
		if curValue.(int8)+addValue.(int8) < maxValue.(int8) {
			return curValue.(int8) + addValue.(int8)
		}
	default:
		panic("Failed to AddGoods. invalid type!")
	}

	return maxValue
}

func getMaxGoodsValue(goodsType int) interface{} {
	switch goodsType {
	case GoodsTypeGold:
		return int64(999999999)
	case GoodsTypeSpin:
		return int16(9999)
	case GoodsTypeShield:
		return int8(3)
	case GoodsTypeCoin:
		return int32(99999999)
	}

	return nil
}

func TimeCompare(time1 time.Time, time2 time.Time, cmptype TIME_TYPE, cmpvalue int, cmpunit TIME_UNIT) bool {
	var cmptime time.Time
	switch cmptype {
	case TIME_COMPARE_OLDER:
		{
			switch cmpunit {
			case TIME_SECOND:
				cmptime = time2.Add(time.Duration(-1*cmpvalue) * time.Second)
			case TIME_MINUTE:
				cmptime = time2.Add(time.Duration(-1*cmpvalue) * time.Minute)
			case TIME_HOUR:
				cmptime = time2.Add(time.Duration(-1*cmpvalue) * time.Hour)
			case TIME_DATE:
				cmptime = time2.AddDate(0, 0, -1*cmpvalue)
			default:
				return false
			}

			if time.Since(time1) >= time.Since(cmptime) {
				return true
			} else {
				return false
			}
		}
	case TIME_COMPARE_EARLYER:
		{
			switch cmpunit {
			case TIME_SECOND:
				cmptime = time2.Add(time.Duration(cmpvalue) * time.Second)
			case TIME_MINUTE:
				cmptime = time2.Add(time.Duration(cmpvalue) * time.Minute)
			case TIME_HOUR:
				cmptime = time2.Add(time.Duration(cmpvalue) * time.Hour)
			case TIME_DATE:
				cmptime = time2.AddDate(0, 0, cmpvalue)
			default:
				return false
			}

			if time.Since(time1) <= time.Since(cmptime) {
				return true
			} else {
				return false
			}
		}
	default:
		{
			return false
		}
	}
}

func KeyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}
