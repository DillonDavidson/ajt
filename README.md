# ajt

Another Japanese Tool - A small CLI tool for AJATT

A collection of handy scripts I put together for managing Japanese media and subtitles more easily.

This CLI combines several tasks I frequently do into one convenient tool.

## Features

- Synchronize subtitle files
- Shift subtitles
- Rename files by extension
- Bulk convert files for passive immersion
- Drop specific audio and subtitle tracks

## Dependencies

- [alass](https://github.com/kaegi/alass) – for synchronizing subtitle files
- [FFmpeg](https://ffmpeg.org/) – extract subtitles
- [impd](https://github.com/Ajatt-Tools/impd) – turn MKV files into OGG files for passive immersion
- [mkvmerge](https://www.matroska.org/downloads/mkvtoolnix.html) – analyze and modify MKV files

## Building

```bash
git clone https://github.com/DillonDavidson/ajt
cd ajt
go install
```

## License

ajt is licensed under the GNU GPL Version 3 license. See [LICENSE](https://github.com/DillonDavidson/ajt/blob/master/LICENSE) for more information.
