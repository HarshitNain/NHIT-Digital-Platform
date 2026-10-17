package utils

import (
	"fmt"
	"math"
	"strings"
)

// CalculateTDS calculates TDS amount based on gross amount and percentage
func CalculateTDS(grossAmount float64, tdsPercentage float64) float64 {
	if tdsPercentage <= 0 {
		return 0
	}
	return math.Round((grossAmount * tdsPercentage / 100) * 100) / 100
}

// CalculateNetPayable calculates net payable amount
// Formula: Gross + Additions - Deductions
func CalculateNetPayable(gross, additions, deductions float64) float64 {
	return math.Round((gross + additions - deductions) * 100) / 100
}

// RoundOff rounds amount to nearest integer
func RoundOff(amount float64) float64 {
	return math.Round(amount)
}

// NumberToWords converts a number to words (Indian numbering system)
func NumberToWords(num float64) string {
	if num == 0 {
		return "Zero only"
	}

	// Split into integer and decimal parts
	intPart := int64(num)
	decimalPart := int64(math.Round((num - float64(intPart)) * 100))

	words := convertToWords(intPart)
	
	if decimalPart > 0 {
		words += " and " + convertToWords(decimalPart) + " Paise"
	}
	
	return words + " only"
}

var ones = []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine"}
var teens = []string{"Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen", "Eighteen", "Nineteen"}
var tens = []string{"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}

func convertToWords(num int64) string {
	if num == 0 {
		return ""
	}

	if num < 10 {
		return ones[num]
	}

	if num < 20 {
		return teens[num-10]
	}

	if num < 100 {
		return tens[num/10] + " " + ones[num%10]
	}

	if num < 1000 {
		return ones[num/100] + " Hundred " + convertToWords(num%100)
	}

	if num < 100000 {
		return convertToWords(num/1000) + " Thousand " + convertToWords(num%1000)
	}

	if num < 10000000 {
		return convertToWords(num/100000) + " Lakh " + convertToWords(num%100000)
	}

	return convertToWords(num/10000000) + " Crore " + convertToWords(num%10000000)
}

// FormatCurrency formats amount as Indian currency
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("₹%.2f", amount)
}

// CleanString removes extra spaces and trims
func CleanString(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}
