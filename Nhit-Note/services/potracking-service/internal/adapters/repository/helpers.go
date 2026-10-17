package repository

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// toPgText converts string to pgtype.Text
func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// toPgTextFromPtr converts *string to pgtype.Text
func toPgTextFromPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// pgTextToPtr converts pgtype.Text to *string
func pgTextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// toPgInt8 converts *int64 to pgtype.Int8
func toPgInt8(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

// toPgInt8FromPtr converts *string to pgtype.Int8
func toPgInt8FromPtr(s *string) pgtype.Int8 {
	if s == nil || *s == "" {
		return pgtype.Int8{Valid: false}
	}
	val, _ := strconv.ParseInt(*s, 10, 64)
	return pgtype.Int8{Int64: val, Valid: true}
}

// pgInt8ToPtr converts pgtype.Int8 to *int64
func pgInt8ToPtr(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

// pgTextToInt64Ptr converts pgtype.Text to *int64
func pgTextToInt64Ptr(t pgtype.Text) *int64 {
	if !t.Valid || t.String == "" {
		return nil
	}
	val, err := strconv.ParseInt(t.String, 10, 64)
	if err != nil {
		return nil
	}
	return &val
}

// toPgInt4 converts *int32 to pgtype.Int4
func toPgInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

// pgInt4ToPtr converts pgtype.Int4 to *int32
func pgInt4ToPtr(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// stringToInt64 converts string to int64
func stringToInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

// stringPtrToString converts *string to string
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// toPgDate converts time.Time to pgtype.Date
func toPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: !t.IsZero()}
}

// toPgNumeric converts float64 to pgtype.Numeric
func toPgNumeric(f float64) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(fmt.Sprintf("%.2f", f))
	return num
}

// parsePrice converts pgtype.Numeric to float64
func parsePrice(num pgtype.Numeric) (float64, error) {
	if !num.Valid {
		return 0, nil
	}
	var f float64
	if num.Int != nil {
		f = float64(num.Int.Int64())
		if num.Exp != 0 {
			for i := int32(0); i < -num.Exp; i++ {
				f /= 10
			}
		}
	}
	return f, nil
}
