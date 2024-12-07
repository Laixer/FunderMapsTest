package utils

import (
	"regexp"
	"strings"
)

type GeocoderDatasource int

const (
	Unknown GeocoderDatasource = iota
	FunderMaps
	FundermapsIncidentReport
	FundermapsInquiryReport
	FundermapsRecoveryReport
	NlPostcode
	NlBagBuilding
	NlBagBerth
	NlBagPosting
	NlBagResidence
	NlBagAddress
	NlCbsNeighborhood
	NlCbsDistrict
	NlCbsMunicipality
	NlCbsState
	NlBagLegacyBuilding
	NlBagLegacyAddress
	NlBagLegacyBerth
	NlBagLegacyPosting
	NlBagLegacyBuildingShort
	NlBagLegacyAddressShort
	NlBagLegacyBerthShort
	NlBagLegacyPostingShort
)

var postcodeRegex = regexp.MustCompile(`^\d{4}[a-zA-Z]{2}$`)

func FromIdentifier(input string) GeocoderDatasource {
	input = strings.ToUpper(strings.ReplaceAll(input, " ", ""))

	switch {
	case strings.HasPrefix(input, "NL.IMBAG.PAND."):
		return NlBagBuilding
	case strings.HasPrefix(input, "NL.IMBAG.LIGPLAATS."):
		return NlBagBerth
	case strings.HasPrefix(input, "NL.IMBAG.STANDPLAATS."):
		return NlBagPosting
	case strings.HasPrefix(input, "NL.IMBAG.VERBLIJFSOBJECT."):
		return NlBagResidence
	case strings.HasPrefix(input, "NL.IMBAG.NUMMERAANDUIDING."):
		return NlBagAddress

	case len(input) == 16 && input[4:6] == "10":
		return NlBagLegacyBuilding
	case len(input) == 16 && input[4:6] == "20":
		return NlBagLegacyAddress
	case len(input) == 16 && input[4:6] == "02":
		return NlBagLegacyBerth
	case len(input) == 16 && input[4:6] == "03":
		return NlBagLegacyPosting

	case len(input) == 15 && input[3:5] == "10":
		return NlBagLegacyBuildingShort
	case len(input) == 15 && input[3:5] == "20":
		return NlBagLegacyAddressShort
	case len(input) == 15 && input[3:5] == "02":
		return NlBagLegacyBerthShort
	case len(input) == 15 && input[3:5] == "03":
		return NlBagLegacyPostingShort

	case strings.HasPrefix(input, "GFM-"):
		return FunderMaps
	case strings.HasPrefix(input, "FIR"):
		return FundermapsIncidentReport
	case strings.HasPrefix(input, "FQR"):
		return FundermapsInquiryReport
	case strings.HasPrefix(input, "FRR"):
		return FundermapsRecoveryReport

	case len(input) == 10 && strings.HasPrefix(input, "BU"):
		return NlCbsNeighborhood
	case len(input) == 8 && strings.HasPrefix(input, "WK"):
		return NlCbsDistrict
	case len(input) == 6 && strings.HasPrefix(input, "GM"):
		return NlCbsMunicipality
	case len(input) == 4 && strings.HasPrefix(input, "PV"):
		return NlCbsState

	case postcodeRegex.MatchString(input):
		return NlPostcode

	default:
		return Unknown
	}
}
