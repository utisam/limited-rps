package lrps

import "testing"

func TestCompareHands(t *testing.T) {
	type args struct {
		a Hand
		b Hand
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Paper vs Rock",
			args: args{a: HandPaper, b: HandRock},
			want: 1,
		},
		{
			name: "Scissor vs Paper",
			args: args{a: HandScissor, b: HandPaper},
			want: 1,
		},
		{
			name: "Rock vs Scissor",
			args: args{a: HandRock, b: HandScissor},
			want: 1,
		},
		{
			name: "Rock vs Paper",
			args: args{a: HandRock, b: HandPaper},
			want: -1,
		},
		{
			name: "Paper vs Scissor",
			args: args{a: HandPaper, b: HandScissor},
			want: -1,
		},
		{
			name: "Scissor vs Rock",
			args: args{a: HandScissor, b: HandRock},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareHands(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("CompareHands() = %v, want %v", got, tt.want)
			}
		})
	}
}
