package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubnet(t *testing.T) {
	dummyHandler := dummyHandler()
	cidr := "192.168.1.0/24"

	testCases := []struct {
		name         string
		ip           string
		expectedCode int
	}{
		{
			name:         "valid",
			ip:           "192.168.1.101",
			expectedCode: http.StatusOK,
		},
		{
			name:         "not trusted",
			ip:           "10.0.0.1",
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(IPHeader, tc.ip)

			_, subnet, err := net.ParseCIDR(cidr)
			require.NoError(t, err)

			wrapper := Subnet(subnet)
			wrapper(dummyHandler).ServeHTTP(rec, req)

			require.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}
