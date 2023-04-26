package addresser

import (
	"reflect"
	"testing"
)

func TestParseAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		want    *Address
		wantErr bool
	}{
		{
			name: "address",
			args: args{
				address: "705 Monterey Pass Rd, Monterey Park, CA 91754",
			},
			want: &Address{
				ID:                "705-Monterey-Pass-Rd%2C-Monterey-Park%2C-CA-91754",
				ZipCode:           "91754",
				ZipCodePlusFour:   "",
				StateAbbreviation: "CA",
				StateName:         "California",
				PlaceName:         "Monterey Park",
				AddressLine1:      "705 Monterey Pass Rd",
				AddressLine2:      "",
				StreetNumber:      "705",
				FormattedAddress:  "705 Monterey Pass Rd, Monterey Park, CA 91754",
				StreetDirection:   "",
				StreetName:        "Monterey Pass Rd",
				StreetSuffix:      "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAddress(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAddress() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
