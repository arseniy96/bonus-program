package converter

import "testing"

func TestConvertFromCent(t *testing.T) {
	type args struct {
		amount int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "valid convert",
			args: args{amount: 399},
			want: 3.99,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertFromCent(tt.args.amount); got != tt.want {
				t.Errorf("ConvertFromCent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToCent(t *testing.T) {
	type args struct {
		amount float64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "valid convert",
			args: args{
				amount: 500.5,
			},
			want: 50050,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertToCent(tt.args.amount); got != tt.want {
				t.Errorf("ConvertToCent() = %v, want %v", got, tt.want)
			}
		})
	}
}
