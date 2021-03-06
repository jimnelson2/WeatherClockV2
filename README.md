# WeatherClockV2
Same, but better. This time with a raspberry pi and Go



## Dev notes

### Build for the pi

```env GOOS=linux GOARCH=arm GOARM=5 go build -o wc-pi ./cmd/weatherclock/weatherclock.go ```

### Install on pi w

* Create a file called wc.json (note, that file is in our .gitignore b/c of the token). It should have this content
  ```json
  {
      "DARKSKY_TOKEN": "DARKSKY_TOKEN_GOES_HERE",
      "LATITUDE": 0.00000,
      "LONGITUDE": -0.00000,
      "DARKSKY_POLL_SEC": 120
  }
  ```

* Copy files to the raspberry pi. Everything below assumes the pi w has ssh enabled already

  * the executable you just built above
 
    ```scp wc-pi pi@raspberrypi.local:~```

  * the fcserver executable
 
    ```scp fcserver-rpi pi@raspberrypi.local:~```

  * the configuration file that has our environment variables
 
    ```scp wc.json pi@raspberrypi.local:~```

  * the service configuration for fcserver
 
    ```scp fcserver.service pi@raspberrypi.local:~```

  * the service configuration for weatherclock
 
    ```scp weatherclock.service pi@raspberrypi.local:~```

  * TODO: the onoff shim stuff...

* move files on the pi w to their correct locations

  * ssh to the pi w
 
    ```ssh pi@raspberrypi.local```

  * move the files (you'll need to be root to do this so...)
    ```bash
    # become root
    sudo -i

    # place the executables. yeah root is gonna run it all. should come back this later...
    mkdir /usr/weatherclock
    mv /home/pi/fcserver-rpi /usr/weatherclock/
    mv /home/pi/wc-pi /usr/weatherclock/
    chmod a+x /usr/weatherclock/*
    chown -R root:root /usr/weatherclock

    # place the configuration file
    mv /home/pi/wc.json /etc/default
    chown -R root:root /etc/default/wc.json
    
    # setup both of our executables to run as services
    mv /home/pi/fcserver.service /etc/systemd/system/
    mv /home/pi/weatherclock.service /etc/systemd/system/
    chown root:root /etc/systemd/system/fcserver.service
    chown root:root /etc/systemd/system/weatherclock.service

    # run the services at startup
    systemctl enable fcserver
    systemctl enable weatherclock
    systemctl start fcserver.service
    systemctl start weatherclock.service
    ```
