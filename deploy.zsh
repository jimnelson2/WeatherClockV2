DEVICE="pi@192.168.127.202"
echo "Building executable"
env GOOS=linux GOARCH=arm GOARM=5 go build -o wc-pi ./cmd/weatherclock/weatherclock.go

echo "copying files"
scp wc-pi ${DEVICE}:~/wc-pi
scp wc.json ${DEVICE}:~/wc.json
scp install.sh ${DEVICE}:~/install.sh

echo "installing"
ssh ${DEVICE} chmod u+x ./install.sh
ssh ${DEVICE} ./install.sh
