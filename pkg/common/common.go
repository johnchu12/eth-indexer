package common

import (
	"encoding/json"
	"math"
	"os"
	"regexp"
	"time"

	"hw/pkg/bigrat"

	"github.com/spf13/cast"
)

// GetEnv retrieves the value of an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func MakeCurrenctTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func MustParseDuration(s string) time.Duration {
	value, err := time.ParseDuration(s)
	if err != nil {
		panic("util: Can't parse duration `" + s + "`: " + err.Error())
	}
	return value
}

func PrintStruct(a any) string {
	data, _ := json.Marshal(a)
	return string(data)
}

func PrintStructRaw(a any) json.RawMessage {
	return json.RawMessage(PrintStruct(a))
}

func GetRegexMap(re *regexp.Regexp, msg string) map[string]string {
	md := make(map[string]string)
	r2 := re.FindAllStringSubmatch(msg, -1)[0]
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			md[name] = r2[i]
		}
	}
	return md
}

// crypto currency pow mining speed format
func FormatHashrate(hashrate string, decimal int32) string {
	zero := math.Log10(cast.ToFloat64(hashrate))
	if zero >= 12 {
		return bigrat.NewBigN(hashrate).Div("1e12").ToTruncateString(decimal) + "T"
	} else if zero >= 9 {
		return bigrat.NewBigN(hashrate).Div("1e9").ToTruncateString(decimal) + "G"
	} else if zero >= 6 {
		return bigrat.NewBigN(hashrate).Div("1e6").ToTruncateString(decimal) + "M"
	} else if zero >= 3 {
		return bigrat.NewBigN(hashrate).Div("1e3").ToTruncateString(decimal) + "K"
	} else {
		return bigrat.NewBigN(hashrate).ToTruncateString(decimal)
	}
}
