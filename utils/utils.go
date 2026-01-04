package utils

import "strings"

// PrependZx adds "0x" prefix if not present
func PrependZx(hex string) string {
    if strings.HasPrefix(hex, "0x") {
        return hex
    }
    return "0x" + hex
}

// RemoveZx removes "0x" prefix if present
func RemoveZx(hex string) string {
    return strings.TrimPrefix(hex, "0x")
}