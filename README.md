# WeatherClockV2
Same, but better. This time with a raspberry pi and Go




## Build for the pi
```env GOOS=linux GOARCH=arm GOARM=5 go build -o wc-pi ./cmd/weatherclock/weatherclock.go ```
