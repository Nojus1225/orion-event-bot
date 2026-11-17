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
