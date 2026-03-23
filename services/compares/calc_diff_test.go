package service_compares

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcDiff(t *testing.T) {
	type args struct {
		local  []string
		remote []string
	}
	tests := []struct {
		name string
		args args
		want []WithDiffState[string]
	}{
		{
			name: "both empty",
			args: args{
				local:  []string{},
				remote: []string{},
			},
			want: []WithDiffState[string]{},
		},
		{
			name: "local only",
			args: args{
				local:  []string{"a", "b"},
				remote: []string{},
			},
			want: []WithDiffState[string]{
				{V: "a", DiffState: LocalOnly},
				{V: "b", DiffState: LocalOnly},
			},
		},
		{
			name: "remote only",
			args: args{
				local:  []string{},
				remote: []string{"a", "b"},
			},
			want: []WithDiffState[string]{
				{V: "a", DiffState: RemoteOnly},
				{V: "b", DiffState: RemoteOnly},
			},
		},
		{
			name: "both",
			args: args{
				local:  []string{"a", "b"},
				remote: []string{"b", "c"},
			},
			want: []WithDiffState[string]{
				{V: "a", DiffState: LocalOnly},
				{V: "c", DiffState: RemoteOnly},
				{V: "b", DiffState: Both},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getId := func(s string) string { return s }
			merge := func(local, remote string) string { return local }

			actual := calcDiff(tt.args.local, tt.args.remote, getId, merge)
			expected := tt.want

			assert.Equal(t, expected, actual)
		})
	}
}
