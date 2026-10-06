package response

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPageResult(t *testing.T) {
	cases := []struct {
		name               string
		total              int64
		page, pageSize     int
		wantPages          int
		wantNext, wantPrev bool
	}{
		{"rỗng", 0, 1, 10, 0, false, false},
		{"một trang", 6, 1, 10, 1, false, false},
		{"trang đầu của nhiều trang", 25, 1, 10, 3, true, false},
		{"trang giữa", 25, 2, 10, 3, true, true},
		{"trang cuối", 25, 3, 10, 3, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPageResult[int](nil, tc.total, tc.page, tc.pageSize)
			assert.Equal(t, tc.wantPages, p.TotalPages)
			assert.Equal(t, tc.wantNext, p.HasNext)
			assert.Equal(t, tc.wantPrev, p.HasPrevious)
			assert.NotNil(t, p.Items, "items rỗng phải serialize thành [] chứ không phải null")
		})
	}
}

func TestBaseResponseJSONShape(t *testing.T) {
	raw, err := json.Marshal(New(200, "ok", NewPageResult([]string{"a"}, 1, 1, 10)))
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))
	assert.ElementsMatch(t, []string{"message", "statusCode", "timestamp", "result"}, keys(got))
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`, got["timestamp"])
	assert.ElementsMatch(t,
		[]string{"items", "total", "page", "pageSize", "totalPages", "hasNext", "hasPrevious"},
		keys(got["result"].(map[string]any)))
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
