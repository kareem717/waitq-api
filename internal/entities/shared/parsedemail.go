package shared

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
)

type ParsedEmail struct {
	Domain string `json:"domain"`
	Local  string `json:"local"`
	Tld    string `json:"tld"`
	Host   string `json:"host"`
}

var _ sql.Scanner = (*ParsedEmail)(nil)

// Scan scans the time parsing it
// i.e (cisco.com,mdencsbo,com,cisco,mdencsbo@cisco.com)
func (pe *ParsedEmail) Scan(src interface{}) (err error) {
	switch src := src.(type) {
	case string:
		// Remove the parentheses
		if len(src) < 2 || src[0] != '(' || src[len(src)-1] != ')' {
			return fmt.Errorf("invalid format for ParsedEmail: %s", src)
		}
		src = src[1 : len(src)-1]

		// Assuming the string format is "domain,localPart,tld,host,plainAddress"
		parts := strings.Split(src, ",")
		if len(parts) != 5 {
			return fmt.Errorf("invalid format for ParsedEmail: %s", src)
		}
		pe.Domain = parts[0]
		pe.Local = parts[1]
		pe.Tld = parts[2]
		pe.Host = parts[3]
		return nil
	default:
		return fmt.Errorf("unsupported data type: %T", src)
	}
}

var _ driver.Valuer = (*ParsedEmail)(nil)

// Value returns the value of the time as a driver.Value.
// i.e (cisco.com,mdencsbo,com,cisco,mdencsbo@cisco.com)
func (pe ParsedEmail) Value() (driver.Value, error) {
	return fmt.Sprintf("(%s,%s,%s,%s)", pe.Domain, pe.Local, pe.Tld, pe.Host), nil
}
