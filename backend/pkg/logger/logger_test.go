package logger_test

import (
	"testing"

	"github.com/finance-sh/finance-sh/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestMaskEmail(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"typical", "diego@finance.sh", "d***@finance.sh"},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"single char local", "a@finance.sh", "a***@finance.sh"},
		{"non-email masked", "justaname", "j***"},
		{"trims surrounding spaces", "  diego@finance.sh  ", "d***@finance.sh"},
		{"leading at is not an email shape", "@finance.sh", "@***"},
		{"multiple at uses last", "weird@name@finance.sh", "w***@finance.sh"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, logger.MaskEmail(tc.in))
		})
	}
}
