package fix

import (
	"fmt"
)

func H(s string, p ...any) []byte {
	return []byte(fmt.Sprintf(s, p...))
}

func headerV0() [][]byte {
	h := make([][]byte, 0, 36)
	h = append(h, H(Mark))
	h = append(h, H(";Header Start"))
	h = append(h, H(";FAVOR:Marlin"))
	h = append(h, H(";TIME:6666"))
	h = append(h, H(";Filament used: %.5fm", Params.AllFilamentUsed()/1000.0))
	h = append(h, H(";Layer height: %.2f", Params.LayerHeight))
	h = append(h, H(";header_type: 3dp"))
	h = append(h, H(";tool_head: %s", Params.ToolHead))
	h = append(h, H(";machine: %s", Params.Model))
	h = append(h, H(";file_total_lines: %d", Params.TotalLines+34))
	h = append(h, H(";estimated_time(s): %.0f", float64(Params.EstimatedTimeSec)*1.07))
	// h = append(h, H(";nozzle_temperature(°C): %.0f", Params.EffectiveNozzleTemperature()))
	h = append(h, H(";nozzle_temperature(°C): %.0f", Params.NozzleTemperatures[0]))
	// h = append(h, H(";nozzle_0_temperature(°C): %.0f", Params.NozzleTemperatures[0]))
	h = append(h, H(";nozzle_0_diameter(mm): %.1f", Params.NozzleDiameters[0]))
	h = append(h, H(";nozzle_0_material: %s", Params.FilamentTypes[0]))
	h = append(h, H(";Extruder 0 Retraction Distance: %.2f", Params.Retractions[0]))
	h = append(h, H(";Extruder 0 Switch Retraction Distance: %.2f", Params.SwitchRetraction[0]))
	h = append(h, H(";nozzle_1_temperature(°C): %.0f", Params.NozzleTemperatures[1]))
	h = append(h, H(";nozzle_1_diameter(mm): %.1f", Params.NozzleDiameters[1]))
	h = append(h, H(";nozzle_1_material: %s", Params.FilamentTypes[1]))
	h = append(h, H(";Extruder 1 Retraction Distance: %.2f", Params.Retractions[1]))
	h = append(h, H(";Extruder 1 Switch Retraction Distance: %.2f", Params.SwitchRetraction[1]))
	h = append(h, H(";build_plate_temperature(°C): %.0f", Params.EffectiveBedTemperature()))
	h = append(h, H(";work_speed(mm/minute): %.0f", Params.PrintSpeedSec*60))
	h = append(h, H(";max_x(mm): %.4f", Params.MaxX))
	h = append(h, H(";max_y(mm): %.4f", Params.MaxY))
	h = append(h, H(";max_z(mm): %.4f", Params.MaxZ))
	h = append(h, H(";min_x(mm): %.4f", Params.MinX))
	h = append(h, H(";min_y(mm): %.4f", Params.MinY))
	h = append(h, H(";min_z(mm): %.4f", Params.MinZ))
	h = append(h, H(";layer_number: %d", Params.TotalLayers))
	h = append(h, H(";layer_height: %.2f", Params.LayerHeight))
	h = append(h, H(";matierial_weight: %.4f", Params.AllFilamentUsedWeight()))
	h = append(h, H(";matierial_length: %.5f", Params.AllFilamentUsed()/1000.0))

	if len(Params.Thumbnail) > 0 {
		h = append(h, H(";thumbnail: %s", Params.Thumbnail))
	}

	h = append(h, H(";Header End\n\n"))
	return h
}

func headerV1() [][]byte {
	h := make([][]byte, 0, 32)
	h = append(h, H(Mark))
	h = append(h, H(";Header Start"))
	h = append(h, H(";Version:1"))
	h = append(h, H(";Printer:%s", Params.Model))
	h = append(h, H(";Estimated Print Time:%d", Params.EstimatedTimeSec))
	h = append(h, H(";Lines:%d", Params.TotalLines+27))
	h = append(h, H(";Extruder Mode:%s", Params.PrintMode))
	h = append(h, H(";Extruder 0 Nozzle Size:%.1f", Params.NozzleDiameters[0]))
	h = append(h, H(";Extruder 0 Material:%s", Params.FilamentTypes[0]))
	h = append(h, H(";Extruder 0 Print Temperature:%.0f", Params.NozzleTemperatures[0]))
	h = append(h, H(";Extruder 0 Retraction Distance:%.2f", Params.Retractions[0]))
	h = append(h, H(";Extruder 0 Switch Retraction Distance:%.2f", Params.SwitchRetraction[0]))
	h = append(h, H(";Extruder 1 Nozzle Size:%.1f", Params.NozzleDiameters[1]))
	h = append(h, H(";Extruder 1 Material:%s", Params.FilamentTypes[1]))
	h = append(h, H(";Extruder 1 Print Temperature:%.0f", Params.NozzleTemperatures[1]))
	h = append(h, H(";Extruder 1 Retraction Distance:%.2f", Params.Retractions[1]))
	h = append(h, H(";Extruder 1 Switch Retraction Distance:%.2f", Params.SwitchRetraction[1]))
	h = append(h, H(";Bed Temperature:%.0f", Params.EffectiveBedTemperature()))
	h = append(h, H(";Work Range - Min X:%.4f", Params.MinX))
	h = append(h, H(";Work Range - Min Y:%.4f", Params.MinY))
	h = append(h, H(";Work Range - Min Z:%.4f", Params.MinZ))
	h = append(h, H(";Work Range - Max X:%.4f", Params.MaxX))
	h = append(h, H(";Work Range - Max Y:%.4f", Params.MaxY))
	h = append(h, H(";Work Range - Max Z:%.4f", Params.MaxZ))

	if Params.LeftExtruderUsed && Params.RightExtruderUsed {
		h = append(h, H(";Extruder(s) Used:2"))
	} else {
		h = append(h, H(";Extruder(s) Used:1"))
	}

	if len(Params.Thumbnail) > 0 {
		h = append(h, H(";Thumbnail:%s", Params.Thumbnail))
	}

	h = append(h, H(";Header End\n\n"))
	return h
}

func ExtractHeader(gcodes []*GcodeBlock) (headers [][]byte, err error) {
	if err = ParseParams(gcodes); err != nil {
		return
	}

	if Params.Version == 1 {
		headers = headerV1()
	} else {
		headers = headerV0()
	}
	return
}
