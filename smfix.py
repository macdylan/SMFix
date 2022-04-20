#!/usr/bin/env python3
"""
Snapmaker2 G-Code Post Processor for PrusaSlicer and SuperSlicer
"""

import re
import sys
from os import getenv

file_input = sys.argv[1]
regex = r'(?:^; thumbnail begin \d+[x ]\d+ \d+)(?:\n|\r\n?)((?:.+(?:\n|\r\n?))+?)(?:^; thumbnail end)'


# slicer_output_name = str(getenv('SLIC3R_PP_OUTPUT_NAME'))
slicer_layer_height = getenv('SLIC3R_LAYER_HEIGHT', 0)
slicer_bed_temperature = getenv('SLIC3R_BED_TEMPERATURE') or getenv('SLIC3R_FIRST_LAYER_BED_TEMPERATURE', 0)
slicer_temperature = getenv('SLIC3R_TEMPERATURE') or getenv('SLIC3R_FIRST_LAYER_TEMPERATURE', 0)
slicer_print_speed_sec = getenv('SLIC3R_MAX_PRINT_SPEED', 0)


def convert_thumbnail(lines):
    comments = ''
    for line in lines:
        if line.startswith(';') or line.startswith('\n'):
            comments += line
    matches = re.findall(regex, comments, re.MULTILINE)
    if len(matches) > 0:
        return 'data:image/png;base64,' + matches[-1:][0].replace('; ', '').replace('\r\n', '').replace('\n', '')
    return None


def find_estimated_time(lines):
    for line in lines:
        if line.startswith('; estimated printing time'):
            est = line[line.index('= ')+2:]  # 2d 12h 8m 58s
            tmp = {'d': 0, 'h': 0, 'm': 0, 's': 0}
            for t in 'dhms':
                if est.find(t) != -1:
                    idx = est.find(t)
                    tmp[t] = int(est[0:idx].replace(' ', ''))
                    est = est[idx+1:]
            return int(tmp['d'] * 86400
                     + tmp['h'] * 3600
                     + tmp['m'] * 60
                     + tmp['s'])


def main():
    with open(file_input, 'r') as f:
        gcode_lines = f.readlines()
        f.close()

    with open(file_input, 'w', newline='') as g:
        thumbnail = convert_thumbnail(gcode_lines)

        headers = (
                ';Header Start',
                ';FAVOR:Marlin',
                ';Layer height: {}'.format(slicer_layer_height),
                ';header_type: 3dp',
                ';thumbnail: {}'.format(thumbnail) if thumbnail else ';',
                ';file_total_lines: {}'.format(len(gcode_lines)),
                ';estimated_time(s): {}'.format(find_estimated_time(gcode_lines)),
                ';nozzle_temperature(°C): {}'.format(slicer_temperature),
                ';build_plate_temperature(°C): {}'.format(slicer_bed_temperature),
                ';work_speed(mm/minute): {}'.format(int(slicer_print_speed_sec) * 60),
                ';Header End\n\n'
                )
        g.write('\n'.join(headers))
        g.writelines(gcode_lines)
        g.close()


if __name__ == "__main__":
    print('Starting SMFix')
    try:
        main()
    except Exception as ex:
        print('Oops! something went wrong.' + str(ex))
        sys.exit(1)
    print('SMFix done')

