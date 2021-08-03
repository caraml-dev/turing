package utils

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

func pad(s string, rightPaddingLength int) string {
	var sb strings.Builder
	for i := 0; i < rightPaddingLength; i++ {
		sb.WriteString("0")
	}

	return fmt.Sprintf("%s%s", s, sb.String())
}

// IsQualifiedKubernetesName allows for padding of name before checking if it is DNS qualified
func IsQualifiedKubernetesName(s string, rightPaddingLength int) bool {
	paddedString := pad(s, rightPaddingLength)
	errs := validation.IsQualifiedName(paddedString)
	return len(errs) == 0
}
