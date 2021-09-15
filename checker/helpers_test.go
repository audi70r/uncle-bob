package checker

import "testing"

func Test_contains(t *testing.T) {
	type args struct {
		s          []string
		searchterm string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "exact match",
			args: args{
				s:          []string{"test", "gagaga", "gigigig"},
				searchterm: "gagaga",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.args.s, tt.args.searchterm); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
