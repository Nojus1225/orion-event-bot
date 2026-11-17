package utils

import "orion-event-bot/bot/vars"

func GetTile(pos int, tiles []vars.PlateCase) vars.PlateCase {
	for _, t := range vars.Plate {
		if t.Number == pos {
			return t
		}
	}

	if len(tiles) == 0 {
		return vars.PlateCase{}
	}

	return tiles[len(tiles)-1]
}
