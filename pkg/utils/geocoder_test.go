package utils

import (
	"testing"
)

func TestFromIdentifier(t *testing.T) {
	testCases := []struct {
		input    string
		expected GeocoderDatasource
	}{
		{"gfm-9d39a41da19044ff80650676981f15fd", FunderMaps},
		{"FIR123", FundermapsIncidentReport},
		{"FQR456", FundermapsInquiryReport},
		{"FRR789", FundermapsRecoveryReport},
		{"1234SB", NlPostcode},
		{"NL.IMBAG.PAND.0301100000028137", NlBagBuilding},
		{"NL.IMBAG.LIGPLAATS.0824030000000238", NlBagBerth},
		{"NL.IMBAG.STANDPLAATS.0629030000033260", NlBagPosting},
		{"NL.IMBAG.VERBLIJFSOBJECT.1676010000517632", NlBagResidence},
		{"NL.IMBAG.NUMMERAANDUIDING.1676200000517717", NlBagAddress},
		{"BU00100203", NlCbsNeighborhood},
		{"WK196609", NlCbsDistrict},
		{"GM0109", NlCbsMunicipality},
		{"PV26", NlCbsState},
		{"1676100000537771", NlBagLegacyBuilding},
		{"0355200000831937", NlBagLegacyAddress},
		// {"123402abcdef", NlBagLegacyBerth},
		// {"123403abcdef", NlBagLegacyPosting},
		{"abcdef", Unknown},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual := FromIdentifier(tc.input)
			if actual != tc.expected {
				t.Errorf("FromIdentifier(%q) = %v; want %v", tc.input, actual, tc.expected)
			}
		})
	}
}
