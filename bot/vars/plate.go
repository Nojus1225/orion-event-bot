package vars

type SpecialCaseType uint8

const (
	SpecialCaseNone SpecialCaseType = iota

	SpecialCaseTree
	SpecialCaseRavine
	SpecialCaseYeti
	SpecialCaseSlush

	SpecialCaseJump
	SpecialCaseBlizzard
	SpecialCaseChairlift

	SpecialCaseIce
	SpecialCaseGondola
	SpecialCaseSupply
)

type PlateCase struct {
	Number      int
	SpecialCase SpecialCaseType
}

var Plate = []PlateCase{
	{Number: 1, SpecialCase: SpecialCaseNone},
	{Number: 2, SpecialCase: SpecialCaseNone},
	{Number: 3, SpecialCase: SpecialCaseNone},
	{Number: 4, SpecialCase: SpecialCaseSupply},
	{Number: 5, SpecialCase: SpecialCaseNone},
	{Number: 6, SpecialCase: SpecialCaseGondola},
	{Number: 7, SpecialCase: SpecialCaseNone},
	{Number: 8, SpecialCase: SpecialCaseIce},
	{Number: 9, SpecialCase: SpecialCaseNone},
	{Number: 10, SpecialCase: SpecialCaseNone},
	{Number: 11, SpecialCase: SpecialCaseTree},
	{Number: 12, SpecialCase: SpecialCaseNone},
	{Number: 13, SpecialCase: SpecialCaseNone},
	{Number: 14, SpecialCase: SpecialCaseNone},
	{Number: 15, SpecialCase: SpecialCaseNone},
	{Number: 16, SpecialCase: SpecialCaseIce},

	{Number: 17, SpecialCase: SpecialCaseNone},
	{Number: 18, SpecialCase: SpecialCaseNone},
	{Number: 19, SpecialCase: SpecialCaseChairlift},
	{Number: 20, SpecialCase: SpecialCaseNone},
	{Number: 21, SpecialCase: SpecialCaseNone},
	{Number: 22, SpecialCase: SpecialCaseSlush},
	{Number: 23, SpecialCase: SpecialCaseNone},
	{Number: 24, SpecialCase: SpecialCaseIce},
	{Number: 25, SpecialCase: SpecialCaseNone},
	{Number: 26, SpecialCase: SpecialCaseBlizzard},
	{Number: 27, SpecialCase: SpecialCaseNone},
	{Number: 28, SpecialCase: SpecialCaseNone},
	{Number: 29, SpecialCase: SpecialCaseSupply},
	{Number: 30, SpecialCase: SpecialCaseNone},
	{Number: 31, SpecialCase: SpecialCaseTree},
	{Number: 32, SpecialCase: SpecialCaseNone},
	{Number: 33, SpecialCase: SpecialCaseIce},
	{Number: 34, SpecialCase: SpecialCaseNone},

	{Number: 35, SpecialCase: SpecialCaseJump},
	{Number: 36, SpecialCase: SpecialCaseNone},
	{Number: 37, SpecialCase: SpecialCaseYeti},
	{Number: 38, SpecialCase: SpecialCaseNone},
	{Number: 39, SpecialCase: SpecialCaseNone},
	{Number: 40, SpecialCase: SpecialCaseIce},
	{Number: 41, SpecialCase: SpecialCaseNone},
	{Number: 42, SpecialCase: SpecialCaseSlush},
	{Number: 43, SpecialCase: SpecialCaseNone},
	{Number: 44, SpecialCase: SpecialCaseTree},
	{Number: 45, SpecialCase: SpecialCaseNone},
	{Number: 46, SpecialCase: SpecialCaseJump},
	{Number: 47, SpecialCase: SpecialCaseNone},
	{Number: 48, SpecialCase: SpecialCaseRavine},
	{Number: 49, SpecialCase: SpecialCaseNone},
	{Number: 50, SpecialCase: SpecialCaseNone},
}
