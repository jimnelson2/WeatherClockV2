
sudo systemctl stop weatherclock.service

sudo mv /home/pi/wc-pi /usr/weatherclock/
sudo chmod a+x /usr/weatherclock/*
sudo chown -R root:root /usr/weatherclock

# place the configuration file
sudo mv /home/pi/wc.json /etc/default
sudo chown -R root:root /etc/default/wc.json

sudo systemctl start weatherclock.service
